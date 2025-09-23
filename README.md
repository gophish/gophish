![mjphish logo](https://github.com/Mauldin-Jenkins/mjphish/blob/89d68bf3a86228db197bcf0234c3d10601c0c101/reports/assets/images/mjlogo.png)

MJphish
=======

![Build Status](https://github.com/gophish/gophish/workflows/CI/badge.svg) [![GoDoc](https://godoc.org/github.com/gophish/gophish?status.svg)](https://godoc.org/github.com/gophish/gophish)

# MJphish: Mauldin & Jenkins Phishing Toolkit

MJphish is a fork of Gophish that has been modified to integrate with the MJSET professional PDF reporting system.

### Install

Details on setting up MJphish on a new system can be found in the [MJSET_INTEGRATION.md](MJSET_INTEGRATION.md) file.

### Setup

Simply run `./mjphish` and navigate to https://localhost:3333 in your browser. MJphish includes user management, so if you need a user account, please ask and one will be created for you.

### Documentation

Documentation for the original Gophish project can be found on our [site](http://getgophish.com/documentation).

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
3. **Docker Support**: Integrate with Docker for easy deployment
4. **Additional Report Types**: Extend with more MJSET report variants
5. **M&J Branding**: Add M&J branding to UI
