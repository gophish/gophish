# MJSET Report Generation - Quick Start Guide

## Overview
The MJSET Report Generation feature allows you to create professional phishing assessment reports directly from the MJphish GUI. Reports are generated in PDF format following Mauldin & Jenkins standards.

## Features
- **GUI-based report generation** with interactive form
- **Customizable report parameters** (test method, impersonated entity, blocking status)
- **Screenshot upload support** for email and landing pages
- **Multiple report types**: Executive Summary, Confidential Report, Internal Report
- **Automatic data extraction** from campaign results
- **Download links** generated automatically after report creation

## How to Use

### Step 1: Start MJphish Server
```bash
cd /Volumes/ADrive/Gits/mjphish
./mjphish
```

### Step 2: Navigate to Campaign Results
1. Open your browser and go to your MJphish instance
2. Log in with your credentials
3. Navigate to **Campaigns** and select a completed campaign
4. Click on the campaign to view results

### Step 3: Generate Reports
1. Scroll to the **MJSET Report Generation** panel
2. Click the **"Generate MJSET Report"** button
3. Fill in the form:
   - **Test Method**: Select from dropdown (e.g., "Malicious link to falsified website")
   - **Impersonated Entity**: Enter who/what was impersonated (e.g., "Microsoft", "CEO")
   - **Was Attack Blocked?**: Select Yes or No
   - **Report Types**: Check which reports to generate (at least one required)
   - **Screenshots** (optional): Upload email and/or landing page screenshots

4. Click **"Generate Reports"** button
5. Wait for the "Reports generated successfully" message
6. Download links will appear automatically

### Step 4: Download Reports
Click on any of the download links to open the PDF in a new tab:
- **Download Executive Report** - High-level summary for executives
- **Download Confidential Report** - Detailed analysis (if selected)
- **Download Internal Report** - Complete technical details (if selected)

## Report Locations
Generated reports are saved to:
```
/Volumes/ADrive/Gits/mjphish/reports/output/campaign_<ID>/
```

Files:
- `SocialEngExecReport.pdf` - Executive Summary
- `Confidential_Report.pdf` - Confidential Report
- `Internal_Use.pdf` - Internal Report

## Command-Line Alternative

You can also generate reports from the command line:

### Basic Usage:
```bash
cd /Volumes/ADrive/Gits/mjphish/reports
python3 gophish_bridge.py <campaign_id>
```

### Interactive Mode:
```bash
python3 gophish_bridge.py <campaign_id> --interactive
```

### With Configuration File:
```bash
python3 gophish_bridge.py <campaign_id> --config /path/to/config.json
```

## Configuration File Format
If using the `--config` option, create a JSON file with:

```json
{
  "campaign_id": 1,
  "test_method": "Malicious link that directed the individual to a falsified website",
  "impersonated_entity": "Microsoft",
  "attack_blocked": "No",
  "report_types": "[\"executive\", \"confidential\"]",
  "email_screenshot": "/path/to/email_screenshot.png",
  "landing_screenshot": "/path/to/landing_screenshot.png",
  "output_dir": "/path/to/output"
}
```

## Troubleshooting

### Reports Not Generating
1. Check that Python 3 is installed: `python3 --version`
2. Verify MJSET dependencies are installed
3. Check the server logs for errors
4. Ensure campaign has completed and has results

### Screenshots Not Appearing
- Screenshots are optional - reports will generate with placeholders if not provided
- Supported formats: PNG, JPG, JPEG, GIF
- Maximum file size: 10MB per image

### Download Links Not Working
1. Ensure reports were successfully generated (check output directory)
2. Verify file permissions on the reports directory
3. Check browser console for JavaScript errors

## API Endpoints

For programmatic access:

### Generate Report
```
POST /api/campaigns/{id}/reports/generate
Content-Type: multipart/form-data
Authorization: Bearer <API_KEY>

Form fields:
- testMethod
- impersonatedEntity
- attackBlocked
- reportTypes (JSON array)
- emailScreenshot (file, optional)
- landingScreenshot (file, optional)
```

### Download Report
```
GET /api/campaigns/{id}/reports/download/{type}
Authorization: Bearer <API_KEY>

Types: executive, confidential, internal
```

### Export Campaign Data
```
GET /api/campaigns/{id}/export
Authorization: Bearer <API_KEY>

Returns JSON with campaign data in MJSET format
```

## Future Enhancements (Optional Go-based Approach)

While the current implementation uses Python/MJSET for PDF generation, a future pure-Go implementation could be added using libraries like:
- `github.com/jung-kurt/gofpdf` - PDF generation
- `github.com/wcharczuk/go-chart` - Chart generation

This would eliminate the Python dependency while maintaining the same functionality.

## Support

For issues or questions:
1. Check the server logs: `tail -f mjphish.log`
2. Check Python script output in reports directory
3. Review campaign data export: `/api/campaigns/{id}/export`

---
**Last Updated:** 2025-10-01
