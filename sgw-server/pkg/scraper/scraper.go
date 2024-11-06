package scraper

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"path"
	"time"

	"github.com/chromedp/chromedp"
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

var (
	ErrPageNotFound     = errors.New("page not found")
	ErrTimeout          = errors.New("operation timed out")
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrRateLimited      = errors.New("rate limited by LinkedIn")
	ErrDataNotFound     = errors.New("required data not found")
	ErrBotDetected      = errors.New("bot detection triggered")
)

func NewScraper(email, password, linkedInURL string) *Scraper {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx)
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)

	s := &Scraper{
		ctx:         ctx,
		cancel:      cancel,
		linkedInURL: linkedInURL,
		email:       email,
		password:    password,
		Profile:     &Profile{},
	}

	s.login()
	return s
}

func (s *Scraper) login() error {
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
	fmt.Println("Logged in...")
	return nil
}

func (s *Scraper) GetRecentPosts() error {
	var posts []Post
	err := chromedp.Run(s.ctx,
		chromedp.Evaluate(`
            try {
                return Array.from(document.querySelectorAll('.feed-shared-update-v2')).slice(0, 5).map(post => {
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
                }).filter(item => item !== null);
            } catch (err) {
                console.error(err);
                return [];
            }
        `, &posts),
	)

	if err != nil {
		return fmt.Errorf("failed to extract posts: %w", err)
	}

	if len(posts) == 0 {
		return ErrDataNotFound
	}

	s.Profile.Posts = posts
	return nil
}

func (s *Scraper) GetExperiences() error {
	url := path.Join(s.linkedInURL, "details/experience")

	// First navigate and scroll
	err := chromedp.Run(s.ctx,
		chromedp.Navigate(url),
		// Add explicit timeout
		chromedp.WaitVisible(`main`, chromedp.ByQuery),

		// Add some random delay to appear more human-like
		chromedp.Sleep(time.Duration(1+rand.Intn(2))*time.Second),

		// Ensure we're logged in by checking for a common LinkedIn element
		chromedp.WaitVisible(`div[data-view-name="profile-component-entity"]`),
	)
	if err != nil {
		return fmt.Errorf("navigation failed: %v", err)
	}
	// Get all experience elements
	//var whatever interface{}
	var experienceElements []Experience
	err = chromedp.Run(s.ctx,
		chromedp.Evaluate(`
        Array.from(document.querySelectorAll('.pvs-list__paged-list-item')).map(el => {
            const position = el.querySelector('div[data-view-name="profile-component-entity"]');
            if (!position) return null;
            
            // Get title from the first level
            const title = position.querySelector('div.display-flex.align-items-center.mr1.t-bold span[aria-hidden="true"]')?.textContent?.trim() 
                        || position.querySelector('div.display-flex.align-items-center.mr1.t-bold span.visually-hidden')?.textContent?.trim()
                        || '';
            
            // Get company from the second level
            const company = position.querySelector('span.t-14.t-normal span[aria-hidden="true"]')?.textContent?.trim() || '';
            
            // Get duration from the third level
            const duration = position.querySelector('span.t-14.t-normal.t-black--light span[aria-hidden="true"]')?.textContent?.trim() || '';

            return { title, company, duration };
        }).filter(item => item !== null);
		`, &experienceElements),
	)

	if err != nil {
		return fmt.Errorf("failed to extract experiences: %v", err)
	}

	return nil
}

func (s *Scraper) GetEducation() error {
	url := path.Join(s.linkedInURL, "details/education")

	err := chromedp.Run(s.ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`main`, chromedp.ByQuery),
		chromedp.Sleep(time.Duration(1+rand.Intn(2))*time.Second),
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
	fmt.Println(educationElements)

	return nil
}

func (s *Scraper) GetNameAndLocation() error {
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
	var about string
	err := chromedp.Run(s.ctx,
		chromedp.Navigate(s.linkedInURL),
		chromedp.Text(`.inline-show-more-text--is-collapsed`, &about),
	)
	if err != nil {
		return fmt.Errorf("failed to get about: %v", err)
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
