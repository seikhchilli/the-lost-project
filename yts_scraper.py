import asyncio
import json
import argparse
from playwright.async_api import async_playwright

async def main(movie_name, match_title=""):
    if not match_title:
        match_title = movie_name
    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        page = await browser.new_page()

        # 1. Navigate to the site
        print("Navigating to https://www13.yts-official.to/ ...")
        await page.goto('https://www13.yts-official.to/', wait_until="domcontentloaded")

        # 2. Type the movie name into search box
        print(f"Searching for '{movie_name}'...")
        # Find search box, usually input[name="keyword"] or #quick-search-input
        search_box = page.locator('input[type="search"], input[name="keyword"], #quick-search-input').first
        await search_box.fill(movie_name)

        # 3. Use send_keys 'Enter'
        print("Sending 'Enter' key...")
        await page.keyboard.press('Enter')

        # 4. Verify search results page
        await page.wait_for_timeout(2000) # Give it 2 seconds for navigation to settle
        await page.wait_for_load_state("domcontentloaded")
        print("On search results page.")

        # 5. Click on the movie titled EXACTLY with the requested name
        print(f"Looking for movie exactly titled '{match_title}'...")
        movie_links = page.locator('a, h2, h3, .movie-title, .title')
        texts = await movie_links.all_text_contents()
        clicked = False
        print(f"Total elements to check: {len(texts)}")
        for i, text in enumerate(texts):
            if text:
                text = text.strip()
                # Case insensitive match allows better matching while being exact in words
                if text.lower() == match_title.lower():
                    print("Found exact match. Clicking...")
                    try:
                        await movie_links.nth(i).click(timeout=3000)
                        clicked = True
                        break
                    except Exception as e:
                        print(f"Failed to click: {e}")

        if not clicked:
            print(f"Could not find exact match '{match_title}'.")
            movie_metadata = {
                "title": match_title,
                "url": page.url,
                "best_quality_link": "",
                "all_links": []
            }
            with open('movie_links.json', 'w') as f:
                json.dump(movie_metadata, f, indent=4)
            await browser.close()
            return

        # 6. Verify movie page
        await page.wait_for_load_state("domcontentloaded")
        print("On movie page.")

        # 7. Extract best quality link
        print("Extracting metadata and best quality link...")
        # Better: just grab all torrent links
        all_links = await page.locator('a').all()
        torrent_links = []
        seen_hrefs = set()
        for a in all_links:
            href = await a.get_attribute('href')
            if href and ('torrent' in href or 'magnet' in href):
                if href not in seen_hrefs:
                    title = await a.get_attribute('title') or await a.inner_text()
                    torrent_links.append({"title": title.strip(), "href": href})
                    seen_hrefs.add(href)

        # Sort by 1080p, 2160p, 4k to find best quality
        best_link = None
        for q in ['2160p', '4K', '1080p', '720p']:
            for t in torrent_links:
                if q.lower() in t['title'].lower():
                    best_link = t['href']
                    break
            if best_link: break

        if not best_link and torrent_links:
            best_link = torrent_links[0]['href']

        movie_metadata = {
            "title": movie_name,
            "url": page.url,
            "best_quality_link": best_link,
            "all_links": torrent_links
        }

        with open('movie_links.json', 'w') as f:
            json.dump(movie_metadata, f, indent=4)

        print("Saved to movie_links.json")
        await browser.close()

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description="Fetch movie from yts")
    parser.add_argument("movie", type=str, help="Name of the movie to search for")
    parser.add_argument("--match", type=str, help="Exact title to match", default="")
    args = parser.parse_args()
    asyncio.run(main(args.movie, args.match))
