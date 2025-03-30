import random
import time
import json
import undetected_chromedriver as uc
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC
from selenium.common.exceptions import TimeoutException, WebDriverException

# ===== USER CONFIG =====
start_page = 1
end_page = 27  # Change this to however many pages you want to scrape
filtered_url = "https://www.dndbeyond.com/magic-items?filter-type=0&filter-search=&filter-requires-attunement=&filter-effect-type=&filter-effect-subtype=&filter-has-charges=&filter-source-category=24&filter-source-category=26&filter-source=146&filter-source=2&filter-source=145&filter-partnered-content=f"

def handle_perimeter_x_challenge(driver):
    """
    Advanced handling of PerimeterX press-and-hold challenge
    """
    try:
        # Wait for challenge container
        challenge_container = WebDriverWait(driver, 10).until(
            EC.presence_of_element_located((By.ID, "px-captcha"))
        )
        
        # Find the iframe within the challenge
        iframe = challenge_container.find_elements(By.TAG_NAME, "iframe")
        
        if not iframe:
            print("No iframe found in challenge")
            return False
        
        # Switch to iframe
        driver.switch_to.frame(iframe[0])
        
        # Find challenge button
        try:
            # Wait for and locate the challenge button
            challenge_button = WebDriverWait(driver, 10).until(
                EC.element_to_be_clickable((By.CSS_SELECTOR, "div[role='button']"))
            )
            
            # Create ActionChains for human-like interaction
            actions = ActionChains(driver)
            
            # Simulate human-like press and hold
            print("Attempting to solve PerimeterX challenge...")
            
            # Move to element first
            actions.move_to_element(challenge_button).perform()
            time.sleep(random.uniform(0.5, 1.5))
            
            # Press and hold with gradual pressure
            actions.click_and_hold(challenge_button)
            
            # Simulate slight movement during hold
            for _ in range(3):
                actions.move_by_offset(
                    random.randint(-5, 5), 
                    random.randint(-5, 5)
                )
                time.sleep(random.uniform(0.1, 0.3))
            
            # Hold for a realistic duration
            time.sleep(random.uniform(1.5, 3))
            
            # Release
            actions.release().perform()
            
            # Additional wait to allow challenge resolution
            time.sleep(random.uniform(1, 2))
            
            # Switch back to default content
            driver.switch_to.default_content()
            
            return True
        
        except Exception as button_error:
            print(f"Error interacting with challenge button: {button_error}")
            # Switch back to default content
            driver.switch_to.default_content()
            return False
    
    except Exception as challenge_error:
        print(f"PerimeterX challenge handling error: {challenge_error}")
        return False

def setup_driver():
    """Advanced driver setup with anti-detection measures"""
    options = uc.ChromeOptions()
    options.add_argument("--disable-blink-features=AutomationControlled")
    options.add_argument("--disable-extensions")
    options.add_argument("--disable-gpu")
    options.add_argument("--no-sandbox")
    options.add_argument("--disable-dev-shm-usage")
    
    # More advanced user agent
    user_agent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
    options.add_argument(f'user-agent={user_agent}')
    
    # Use undetected_chromedriver to bypass bot detection
    driver = uc.Chrome(options=options)
    
    # Additional browser configurations
    driver.set_page_load_timeout(30)
    driver.implicitly_wait(10)
    
    return driver

def scrape_magic_items(start_page=1, end_page=27):
    driver = setup_driver()
    all_items = []
    error_pages = []

    try:
        for page in range(start_page, end_page + 1):
            url = f"{filtered_url}&page={page}"
            print(f"\n>>> Scraping Page {page} | URL: {url}")
            
            try:
                driver.get(url)
                
                # Check and handle PerimeterX challenge
                if not handle_perimeter_x_challenge(driver):
                    print(f"❗ Failed to bypass challenge on page {page}")
                    error_pages.append((page, url))
                    continue
                
                # Wait for items to load
                items = WebDriverWait(driver, 15).until(
                    EC.presence_of_all_elements_located((By.CSS_SELECTOR, "span.name a.link"))
                )
                
                for idx, item in enumerate(items, start=1):
                    try:
                        title = item.text.strip()
                        link = item.get_attribute('href')
                        print(f"Page {page} - Item {idx}: {title} | {link}")
                        all_items.append({'title': title, 'link': link})
                        
                        # Random delay between item parsing
                        time.sleep(random.uniform(0.5, 1.5))
                    except Exception as e:
                        print(f"❗️ Error reading item {idx} on page {page}: {e}")
                
                # Random delay between pages
                time.sleep(random.uniform(3, 6))
                
            except Exception as page_error:
                print(f"❗ Failed to scrape page {page}: {page_error}")
                error_pages.append((page, url))
                continue

    except WebDriverException as e:
        print(f"❗️ Critical driver error: {e}")
    
    finally:
        driver.quit()

        # Save results
        with open('magic_items.json', 'w', encoding='utf-8') as f:
            json.dump(all_items, f, ensure_ascii=False, indent=2)

        # ===== Summary =====
        print("\n" + "-" * 60)
        print(f"✅ Total Items Scraped: {len(all_items)}")
        if error_pages:
            print(f"❗️ Pages failed to scrape: {len(error_pages)}")
            for ep in error_pages:
                print(f" - Page {ep[0]}: {ep[1]}")
        else:
            print("🎉 All pages scraped successfully!")

        return all_items, error_pages

# Run the scraper
if __name__ == "__main__":
    scrape_magic_items(start_page, end_page)