package scraper

import (
	"bufio"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"os"
	"path"
	"strings"
	"time"
)

type Experience struct {
	Company  string `json:"company"`
	Duration string `json:"duration"`
	Title    string `json:"title"`
}

type Post struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Education struct {
	Institute string `json:"institute"`
	Major     string `json:"major"`
	Duration  string `json:"duration"`
}

type Profile struct {
	Name       string
	Location   string
	About      string
	Experience []Experience
	Education  []Education
	Posts      []Post
}

type Scraper struct {
	ctx         context.Context
	cancel      context.CancelFunc
	linkedInURL string
	email       string
	password    string
	Profile     *Profile
}

func NewScraper(email, password, linkedInURL string) (*Scraper, error) {

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // Start headless
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("disable-setuid-sandbox", true),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx)
	ctx, cancel = context.WithTimeout(ctx, 3*time.Minute)
	s := &Scraper{
		ctx:         ctx,
		cancel:      cancel,
		linkedInURL: linkedInURL,
		email:       email,
		password:    password,
		Profile:     &Profile{},
	}

	err := s.login(false)
	if err == nil {
		return s, nil
	}

	// If we get to a verification page, restart with visible browser
	if strings.Contains(err.Error(), "verification") {
		s.cancel() // Clean up the first browser

		// Create visible browser for verification
		visibleOpts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", false),
			chromedp.Flag("disable-gpu", false),
			chromedp.Flag("disable-extensions", false),
			chromedp.Flag("disable-setuid-sandbox", true),
		)
		visibleAllocCtx, visibleCancel := chromedp.NewExecAllocator(context.Background(), visibleOpts...)
		defer visibleCancel()

		ctx, _ = chromedp.NewContext(visibleAllocCtx)
		ctx, cancel = context.WithTimeout(ctx, 3*time.Minute)
		s.ctx = ctx
		s.cancel = cancel

		// Try login with visible browser
		if err := s.login(false); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to login even with verification: %w", err)
		}

		// After successful verification, switch back to headless
		s.cancel()
		ctx, _ = chromedp.NewContext(allocCtx)
		ctx, cancel = context.WithTimeout(ctx, 3*time.Minute)
		s.ctx = ctx
		s.cancel = cancel
	} else {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return s, nil
}

func (s *Scraper) login(headless bool) error {
	fmt.Println("Logging user in...")

	err := chromedp.Run(s.ctx,
		chromedp.Navigate("https://www.linkedin.com/login"),
		chromedp.WaitVisible(`input[name="session_key"]`),
		chromedp.SendKeys(`input[name="session_key"]`, s.email),
		chromedp.SendKeys(`input[name="session_password"]`, s.password),
		chromedp.Click(`button[type="submit"]`),
	)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)

	var currentURL string
	err = chromedp.Run(s.ctx,
		chromedp.Location(&currentURL),
	)
	if err != nil {
		return err
	}

	if strings.Contains(currentURL, "checkpoint/challenge") {
		if headless {
			return fmt.Errorf("verification required, please retry with headless=false")
		}

		fmt.Println("\nSecurity verification required!")
		fmt.Println("Please complete the verification puzzle in the browser window")
		fmt.Print("\nPress Enter once you've completed the verification...")
		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadString('\n')

		err = chromedp.Run(s.ctx,
			chromedp.Location(&currentURL),
		)
		if err != nil {
			return err
		}
		if strings.Contains(currentURL, "checkpoint/challenge") {
			return fmt.Errorf("verification was not completed successfully")
		}
	}

	fmt.Println("Logged in successfully")
	return nil
}

func (s *Scraper) GetRecentPosts() error {
	fmt.Println("Getting latest posts")
	url := path.Join(s.linkedInURL, "recent-activity/all/")
	var posts []Post
	err := chromedp.Run(s.ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`
                 Array.from(document.querySelectorAll('.feed-shared-update-v2')).map(post => {
                    // Check if it's a repost by looking for specific class or text in header
                    const header = post.querySelector('.update-components-header__text-view');
                    if (header && header.textContent.includes('reposted this')) {
                        return null;
                    }

                    // Get the content wrapper
                    const wrapper = post.querySelector('.feed-shared-update-v2__description-wrapper');
                    if (!wrapper) return null;

                    const title = wrapper.querySelector('.break-words span[dir="ltr"]')?.textContent?.trim() || '';
                    const content = wrapper.querySelector('.feed-shared-inline-show-more-text')?.textContent?.trim() || '';
                    
                    if (!title && !content) return null;

                    return {
                        title: title,
                        content: content
                    };
                }).filter(item => item !== null).slice(0, 5);
        `, &posts),
	)

	fmt.Println(posts)
	fmt.Println()

	if err != nil {
		return fmt.Errorf("failed to extract posts: %w", err)
	}

	s.Profile.Posts = posts
	return nil
}

func (s *Scraper) GetExperiences() error {
	fmt.Println("Getting experience")
	url := path.Join(s.linkedInURL, "details/experience")

	err := chromedp.Run(s.ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`main`, chromedp.ByQuery),
		chromedp.WaitVisible(`div[data-view-name="profile-component-entity"]`),
	)
	if err != nil {
		return fmt.Errorf("navigation failed: %v", err)
	}

	var experienceElements []Experience
	err = chromedp.Run(s.ctx,
		chromedp.Evaluate(`
        Array.from(document.querySelectorAll('.pvs-list__paged-list-item')).map(el => {
            const position = el.querySelector('div[data-view-name="profile-component-entity"]');
            if (!position) return null;
            const title = position.querySelector('div.display-flex.align-items-center.mr1.t-bold span[aria-hidden="true"]')?.textContent?.trim() 
                        || position.querySelector('div.display-flex.align-items-center.mr1.t-bold span.visually-hidden')?.textContent?.trim()
                        || '';
            const company = position.querySelector('span.t-14.t-normal span[aria-hidden="true"]')?.textContent?.trim() || '';
            const duration = position.querySelector('span.t-14.t-normal.t-black--light span[aria-hidden="true"]')?.textContent?.trim() || '';
            return { title, company, duration };
        }).filter(item => item !== null);
		`, &experienceElements),
	)

	if err != nil {
		return fmt.Errorf("failed to extract experiences: %v", err)
	}

	s.Profile.Experience = experienceElements

	return nil
}

func (s *Scraper) GetEducation() error {
	fmt.Println("Getting education")
	url := path.Join(s.linkedInURL, "details/education")

	err := chromedp.Run(s.ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`main`, chromedp.ByQuery),
		chromedp.WaitVisible(`div[data-view-name="profile-component-entity"]`),
	)
	if err != nil {
		return fmt.Errorf("navigation failed: %v", err)
	}

	var educationElements []Education
	err = chromedp.Run(s.ctx,
		chromedp.Evaluate(`
            Array.from(document.querySelectorAll('.pvs-list__paged-list-item')).map(el => {
                const position = el.querySelector('div[data-view-name="profile-component-entity"]');
                if (!position) return null;
                const institute = position.querySelector('div.display-flex.align-items-center.mr1.hoverable-link-text.t-bold span[aria-hidden="true"]')?.textContent?.trim() || '';
                const major = position.querySelector('span.t-14.t-normal span[aria-hidden="true"]')?.textContent?.trim() || '';
                const duration = position.querySelector('span.t-14.t-normal.t-black--light span[aria-hidden="true"]')?.textContent?.trim() || '';
                return {
                    institute,
                    major,
                    duration
                };
            }).filter(item => item !== null);
		`, &educationElements),
	)
	if err != nil {
		return fmt.Errorf("failed to extract education: %v", err)
	}
	s.Profile.Education = educationElements

	return nil
}

func (s *Scraper) GetNameAndLocation() error {
	fmt.Println("Getting name and location")
	var name, location string
	err := chromedp.Run(s.ctx,
		chromedp.Navigate(s.linkedInURL),
		chromedp.WaitVisible(`.mt2.relative`),
		chromedp.Text(`h1.inline.t-24.v-align-middle.break-words`, &name),
		chromedp.Text(`.text-body-small.inline.t-black--light.break-words`, &location),
	)
	if err != nil {
		return fmt.Errorf("failed to get name and location: %v", err)
	}

	s.Profile.Name = name
	s.Profile.Location = location
	return nil
}

func (s *Scraper) GetAbout() error {
	fmt.Println("Getting about")
	var about string
	err := chromedp.Run(s.ctx,
		chromedp.WaitVisible(`div[class*="display-flex ph5"]`), // Wait for main content
		chromedp.Evaluate(`(() => {
            // Find the About section's text content
            const aboutSpans = document.querySelectorAll('div[class*="display-flex full-width"] span[aria-hidden="true"]');
            if (!aboutSpans.length) return "";
            
            return Array.from(aboutSpans)
                .map(span => span.textContent.trim())
                .filter(text => text.length > 0)[0]
        })()`, &about),
	)

	if err != nil {
		return fmt.Errorf("failed to get about: %w", err)
	}

	s.Profile.About = about
	return nil
}

func (s *Scraper) Close() {
	s.cancel()
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
