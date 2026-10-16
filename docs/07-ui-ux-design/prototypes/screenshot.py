#!/usr/bin/env python3
"""
Generate screenshots from HTML design files using Playwright.

This script opens each HTML file in a headless browser and takes a light
(`neo-light`) and a dark (`neo-dark`) full-page screenshot. Screenshots are
written to `images/light/` and `images/dark/` using the HTML file's name.

Usage:
    uv run screenshot.py

Or with Python directly:
    python3 screenshot.py

First-time setup:
    uv sync
    uv run playwright install chromium
"""

import asyncio
from pathlib import Path
from playwright.async_api import async_playwright

html_dir = Path(__file__).parent / "html"
images_dir = Path(__file__).parent / "images"

# (theme name stored in localStorage, output subdirectory)
THEMES = (("neo-light", "light"), ("neo-dark", "dark"))


async def main():
    # Ensure image directories exist
    for _, subdir in THEMES:
        (images_dir / subdir).mkdir(parents=True, exist_ok=True)

    # Get all HTML files sorted by name
    html_files = sorted(html_dir.glob("*.html"))

    if not html_files:
        print(f"No HTML files found in {html_dir}")
        return

    print(f"Found {len(html_files)} HTML files to screenshot")

    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        context = await browser.new_context(
            viewport={'width': 1920, 'height': 1080}
        )
        page = await context.new_page()

        for html_file in html_files:
            file_url = f"file://{html_file.absolute()}"

            for theme, subdir in THEMES:
                png_path = images_dir / subdir / f"{html_file.stem}.png"

                print(f"Screenshotting {html_file.name} [{theme}] -> {subdir}/{png_path.name}")

                try:
                    await page.goto(file_url, wait_until="networkidle")
                    await page.evaluate(
                        "theme => window.localStorage.setItem('neo-theme', theme)",
                        theme,
                    )
                    await page.reload(wait_until="networkidle")
                    await page.wait_for_timeout(500)  # Wait for any animations

                    # Take full page screenshot
                    await page.screenshot(path=str(png_path), full_page=True)
                    print(f"  ✓ Saved to {png_path}")
                except Exception as e:
                    print(f"  ✗ Failed: {e}")

        await browser.close()

    print(f"\n{len(html_files)} HTML files captured in {len(THEMES)} themes")


if __name__ == "__main__":
    asyncio.run(main())
