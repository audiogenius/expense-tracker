Google Cloud Vision API - Step by step (for beginners)

1. Create Google account (if not already): https://accounts.google.com/
2. Go to Google Cloud Console: https://console.cloud.google.com/
3. Create a new Project: Top-left -> Select a project -> New Project -> give a name -> Create
4. Enable Vision API: Navigation Menu -> APIs & Services -> Library -> search "Vision API" -> Enable
5. Create API Key: APIs & Services -> Credentials -> + CREATE CREDENTIALS -> API key
   - Copy the API key and store it in your `.env` as GCV_API_KEY=YOUR_KEY
6. (Optional) Set billing to get higher quota: Billing -> Link a billing account

Notes:
- Free tier: up to certain requests per month.
- Make sure to restrict API key for security: Click the created key -> Application restrictions -> HTTP referrers or IP addresses -> fill accordingly.

Where to put in `.env`:
GCV_API_KEY=your_google_cloud_vision_api_key
USE_LOCAL_OCR=false

This project uses GCV as primary OCR and local Tesseract as fallback.
