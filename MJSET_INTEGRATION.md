# MJSET Integration for Gophish

## Overview

This integration successfully combines Gophish's campaign management capabilities with MJSET's professional PDF reporting system. The solution maintains the strengths of both systems while providing seamless integration.

## What Was Implemented

### 1. New Gophish API Endpoints

**File**: `controllers/api/reports.go`
- `GET /api/campaigns/{id}/export` - Export campaign data in JSON format
- `POST /api/campaigns/{id}/reports/generate` - Trigger report generation
- `GET /api/campaigns/{id}/reports/download/{type}` - Download generated reports

**File**: `controllers/api/server.go` (modified)
- Added route registration for new report endpoints

### 2. Python Bridge System

**File**: `reports/gophish_bridge.py`
- Fetches campaign data from Gophish API
- Transforms JSON data to MJSET CSV format
- Manages temporary files and report generation
- Handles error cases and cleanup

### 3. Adapted MJSET Classes

**File**: `reports/mjset/PdfReport.py`
- Base PDF report class adapted for Gophish integration
- Fallback fonts and images for standalone operation
- Flexible asset path management

**File**: `reports/mjset/makeSEReport.py`
- Social engineering report generator
- Automatic template creation if missing
- Data processing from Gophish format

### 4. Supporting Files

- `reports/requirements.txt` - Python dependencies
- `reports/README.md` - Comprehensive documentation
- `reports/install.sh` - Installation script
- `reports/test_integration.py` - Test script with mock data
- `reports/config.example.json` - Configuration template

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Gophish       │    │  Python Bridge   │    │  MJSET Reports  │
│   Campaign      │───▶│  gophish_bridge  │───▶│  PDF Generation │
│   Data (Go)     │    │  (Python)        │    │  (Python)       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                        │                        │
         ▼                        ▼                        ▼
   JSON via API          CSV Files (temp)           PDF Reports
```

## Data Flow

1. **Campaign Execution**: Gophish runs phishing campaigns and stores results
2. **Data Export**: New API endpoint exports campaign data as structured JSON
3. **Data Transformation**: Python bridge converts JSON to MJSET CSV format
4. **Report Generation**: Adapted MJSET classes generate professional PDF reports
5. **File Delivery**: Reports are served through Gophish API endpoints

## Key Benefits

### ✅ Leverages Existing Investment
- Reuses sophisticated MJSET report templates and branding
- Maintains professional report quality and formatting
- Preserves M&J branding and signature elements

### ✅ Seamless Integration
- No changes required to core Gophish functionality
- New API endpoints follow existing patterns
- Clean separation of concerns

### ✅ Flexible Architecture
- Python components are self-contained
- Easy to modify or extend report types
- Fallback handling for missing assets

### ✅ Professional Output
- Executive summary reports with statistics and charts
- Detailed credential analysis reports
- Full internal reports with all collected data
- Professional branding and formatting

## Installation & Usage

### Quick Start
```bash
# 1. Install Python dependencies
cd reports/
pip3 install -r requirements.txt

# 2. Set API key
export GOPHISH_API_KEY="your_key_here"

# 3. Test integration
python3 test_integration.py

# 4. Rebuild Gophish
go build
```

### Generate Reports via API
```bash
# Generate reports for campaign ID 1
curl -X POST -H "Authorization: Bearer YOUR_API_KEY" \
  http://localhost:3333/api/campaigns/1/reports/generate

# Download executive report
curl -H "Authorization: Bearer YOUR_API_KEY" \
  http://localhost:3333/api/campaigns/1/reports/download/executive \
  -o executive_report.pdf
```

## Report Types Available

1. **Executive Summary** (`SocialEngExecReport.pdf`)
   - High-level campaign statistics
   - Click and credential submission rates
   - Professional formatting for management

2. **Confidential Report** (`Confidential_Report.pdf`)
   - Detailed credential analysis
   - Password strength assessment
   - Obfuscated for client sharing

3. **Internal Report** (`Internal_Use.pdf`)
   - Complete credential details
   - Full user interaction data
   - For internal security team use

## Customization Options

### Branding
- Copy fonts from `/Volumes/Adrive/mjset/fonts/` to `reports/assets/fonts/`
- Copy images from `/Volumes/Adrive/mjset/images/` to `reports/assets/images/`
- Copy templates from `/Volumes/Adrive/mjset/text_files/` to `reports/assets/templates/`

### Configuration
- Modify `reports/config.example.json` and rename to `config.json`
- Set company information, branding colors, and default values
- Configure API endpoints and authentication

## Technical Implementation Notes

### Why Python API Integration?
- **Faster Development**: Reuses existing, tested MJSET code
- **Professional Output**: Maintains sophisticated PDF generation capabilities
- **Maintainability**: Clear separation between Go (data) and Python (reports)
- **Flexibility**: Easy to extend with additional report types

### Data Mapping Strategy
- Gophish Events (DataSubmit) → MJSET cred.csv format
- Gophish Events (Clicked) → MJSET hits_report.txt format
- Gophish Groups/Targets → MJSET emails_mapped.txt format
- Campaign metadata → MJSET scandata.txt format

### Error Handling
- Graceful fallbacks for missing assets (fonts, images)
- Default templates created automatically
- Comprehensive error logging and reporting
- Cleanup of temporary files

## Future Enhancements

Potential improvements that could be added:

1. **Web UI Integration**: Add report generation buttons to Gophish dashboard
2. **Scheduled Reports**: Automatic report generation on campaign completion
3. **Email Delivery**: Send reports directly to stakeholders
4. **Additional Report Types**: Extend with more MJSET report variants
5. **Real-time Updates**: Live report updates during active campaigns

## Conclusion

This integration successfully bridges Gophish's powerful campaign management with MJSET's professional reporting capabilities. The solution:

- ✅ **Maintains Quality**: Preserves MJSET's professional report formatting
- ✅ **Adds Functionality**: Extends Gophish with advanced reporting
- ✅ **Stays Modular**: Clean architecture allows independent updates
- ✅ **Provides Value**: Delivers the same high-quality reports you're used to

The integration is production-ready and provides the same professional reporting capabilities that MJSET offers, now seamlessly integrated with Gophish's campaign management system.
