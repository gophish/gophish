# Gophish-MJSET Report Integration

This directory contains the integration between Gophish and MJSET reporting capabilities, allowing you to generate professional PDF reports from Gophish campaign data.

## Overview

The integration provides:
- **Executive Summary Reports**: Professional PDF reports with campaign statistics, charts, and branding
- **API Integration**: New Gophish API endpoints for data export and report generation
- **Automated Processing**: Converts Gophish data to MJSET format automatically

## Architecture

```
Gophish Campaign Data → API Export → Python Bridge → MJSET Reports → PDF Files
```

## Installation

1. **Install Python Dependencies**:
   ```bash
   cd reports/
   pip3 install -r requirements.txt
   ```

2. **Set up Assets** (Optional):
   - Copy fonts, images, and templates from your original MJSET installation to `reports/assets/`
   - The system will use fallbacks if assets are not available

## API Endpoints

The integration adds these new endpoints to Gophish:

### Data Export
- `GET /api/campaigns/{id}/export` - Export campaign data in JSON format

### Report Generation  
- `POST /api/campaigns/{id}/reports/generate` - Generate all report types
- `GET /api/campaigns/{id}/reports/download/{type}` - Download specific report

Report types:
- `executive` - Executive summary report (SocialEngExecReport.pdf)
- `confidential` - Detailed credential report (Confidential_Report.pdf) 
- `internal` - Full internal report (Internal_Use.pdf)

## Usage

### Via API

1. **Generate Reports**:
   ```bash
   curl -X POST -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:3333/api/campaigns/1/reports/generate
   ```

2. **Download Report**:
   ```bash
   curl -H "Authorization: Bearer YOUR_API_KEY" \
     http://localhost:3333/api/campaigns/1/reports/download/executive \
     -o executive_report.pdf
   ```

### Via Command Line

You can also generate reports directly:

```bash
cd reports/
python3 gophish_bridge.py CAMPAIGN_ID
```

## Testing

Run the test script to verify the integration:

```bash
cd reports/
python3 test_integration.py
```

This will generate sample reports using mock data.

## File Structure

```
reports/
├── README.md                 # This file
├── requirements.txt          # Python dependencies
├── gophish_bridge.py        # Main bridge script
├── test_integration.py      # Test script
├── mjset/                   # Adapted MJSET classes
│   ├── __init__.py
│   ├── PdfReport.py        # Base PDF report class
│   └── makeSEReport.py     # Social engineering report class
├── assets/                  # Fonts, images, templates (optional)
│   ├── fonts/
│   ├── images/
│   └── templates/
└── output/                  # Generated reports
    └── campaign_X/         # Reports for campaign X
```

## Configuration

### API Key

The bridge script needs a Gophish API key. Set it via:

1. **Environment Variable**:
   ```bash
   export GOPHISH_API_KEY="your_api_key_here"
   ```

2. **Config File**: The script will read from `../config.json` if available

### Customization

You can customize reports by:

1. **Templates**: Modify files in `assets/templates/`
2. **Branding**: Replace images in `assets/images/`
3. **Fonts**: Add custom fonts to `assets/fonts/`

## Data Mapping

The integration maps Gophish data to MJSET format:

| Gophish | MJSET File | Purpose |
|---------|------------|---------|
| Campaign Results → Events (DataSubmit) | cred.csv | Credential submissions |
| Campaign Events (Clicked) | hits_report.txt | Click tracking |
| Campaign Groups → Targets | emails_mapped.txt | Target email list |
| Campaign Metadata | scandata.txt | Campaign information |

## Troubleshooting

### Common Issues

1. **"Report generation script not found"**
   - Ensure `gophish_bridge.py` is executable: `chmod +x reports/gophish_bridge.py`

2. **"Module not found" errors**
   - Install dependencies: `pip3 install -r requirements.txt`
   - Check Python path in the bridge script

3. **"Campaign not found"**
   - Verify the campaign ID exists and you have access
   - Check API key permissions

4. **Font/Image errors**
   - Reports will use fallbacks if assets are missing
   - Copy assets from original MJSET installation for full branding

### Debug Mode

Enable debug output:
```bash
export GOPHISH_DEBUG=1
python3 gophish_bridge.py CAMPAIGN_ID
```

## Security Notes

- API keys should be kept secure and rotated regularly
- Generated reports may contain sensitive credential data
- Ensure proper file permissions on the output directory
- Consider encrypting report files for sensitive assessments

## Integration with Original MJSET

If you have an existing MJSET installation:

1. Copy assets: `cp -r /path/to/mjset/{fonts,images,templates} reports/assets/`
2. Update paths in `PdfReport.py` if needed
3. Test with your existing templates and branding

## Support

For issues with:
- **Gophish Integration**: Check the API endpoints and Go code
- **Report Generation**: Check the Python bridge and MJSET classes  
- **PDF Output**: Verify ReportLab installation and template files
