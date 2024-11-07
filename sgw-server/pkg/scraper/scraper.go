/*
	Package scraper provides functionality to programmatically extract profile information from LinkedIn.

It uses Chrome DevTools Protocol (CDP) via the chromedp package to automate browser interactions
and extract various sections of LinkedIn profiles including basic information, experience,
education, and recent posts.
Scraping is down by injecting javscript in the launched chrome instance, and getting the results

Basic usage:

	scraper, err := scraper.NewScraper("email", "password", "https://www.linkedin.com/in/username")
	if err != nil {
	    log.Fatal(err)
	}
	defer scraper.Close()

Fetch profile information:

	scraper.GetNameAndLocation()
	scraper.GetAbout()
	scraper.GetExperiences()
	scraper.GetEducation()
	scraper.GetRecentPosts()
*/
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

/*
	Experience represents a single work experience entry from a LinkedIn profile.

It contains details about the job title, company, and duration of employment.
*/
type Experience struct {
	Company  string `json:"company"`  // Name of the employer
	Duration string `json:"duration"` // Period of employment (e.g., "2019 - Present")
	Title    string `json:"title"`    // Job title or role
}

/*
	Post represents a LinkedIn post made by the profile owner.

It contains the textual content of the post.
*/
type Post struct {
	Content string `json:"content"` // Text content of the post
}

/*
	Education represents an educational background entry from a LinkedIn profile.

It contains details about the educational institution, field of study, and duration.
*/
type Education struct {
	Institute string `json:"institute"` // Name of the educational institution
	Major     string `json:"major"`     // Field of study or degree program
	Duration  string `json:"duration"`  // Period of study (e.g., "2015 - 2019")
}

/*
	Profile represents the complete LinkedIn profile information that can be scraped.

It contains all the profile sections including personal info, experiences,
education history, and recent posts.
*/
type Profile struct {
	Name       string       // Full name of the profile owner
	Location   string       // Geographic location
	About      string       // "About" section content
	Experience []Experience // List of work experiences
	Education  []Education  // List of education entries
	Posts      []Post       // List of recent posts
}

/*
	Scraper handles the LinkedIn profile scraping operations.

It maintains the browser context and authentication state required
for accessing LinkedIn profile information.
*/
type Scraper struct {
	ctx         context.Context
	cancel      context.CancelFunc
	linkedInURL string
	email       string
	password    string
	Profile     *Profile
}

/*
	NewScraper creates and initializes a new LinkedIn scraper with the provided credentials.

It handles the initial login process and automatically manages browser visibility
for security verification if required.

Parameters:
  - email: LinkedIn account email
  - password: LinkedIn account password
  - linkedInURL: Target profile URL to scrape

Returns:
  - *Scraper: Initialized scraper instance
  - error: Any error encountered during setup or login
*/
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

/*
	login authenticates with LinkedIn using the provided credentials.

It automatically handles security verification challenges by switching
to a visible browser when necessary.

Parameters:
  - headless: Whether to use headless browser mode

Returns:
  - error: Any error encountered during login
*/
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

/*
	GetRecentPosts retrieves the 5 most recent posts from the profile,

excluding reposts. The results are stored in Profile.Posts.

Returns:
  - error: Any error encountered while fetching posts
*/
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
                    const content = wrapper.querySelector('.feed-shared-inline-show-more-text')?.textContent?.trim() || wrapper.querySelector('.break-words span[dir="ltr"]')?.textContent?.trim() || '';;
                    
                    if (!content) return null;

                    return {
                        content: content
                    };
                }).filter(item => item !== null).slice(0, 5);
        `, &posts),
	)

	if err != nil {
		return fmt.Errorf("failed to extract posts: %w", err)
	}

	s.Profile.Posts = posts
	return nil
}

/*
	GetExperiences extracts work experience entries from the profile.

The results are stored in Profile.Experience.

Returns:
  - error: Any error encountered while fetching experiences
*/
func (s *Scraper) GetExperiences() error {
	fmt.Println("Getting experience")
	url := path.Join(s.linkedInURL, "details/experience")

	err := chromedp.Run(s.ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
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

/*
	GetEducation extracts education history from the profile.

The results are stored in Profile.Education.

Returns:
  - error: Any error encountered while fetching education
*/
func (s *Scraper) GetEducation() error {
	fmt.Println("Getting education")
	url := path.Join(s.linkedInURL, "details/education")

	err := chromedp.Run(s.ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
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

/*
	GetNameAndLocation retrieves the profile owner's name and location.

The results are stored in Profile.Name and Profile.Location.

Returns:
  - error: Any error encountered while fetching name and location
*/
func (s *Scraper) GetNameAndLocation() error {
	fmt.Println("Getting name and location")
	var name, location string
	err := chromedp.Run(s.ctx,
		chromedp.Navigate(s.linkedInURL),
		chromedp.Sleep(2*time.Second),
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

/*
	GetAbout extracts the "About" section content from the profile.

The result is stored in Profile.About.

Returns:
  - error: Any error encountered while fetching about section
*/
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

/*
	Close releases all resources associated with the scraper,

including the browser context. This should be called when
the scraper is no longer needed.
*/
