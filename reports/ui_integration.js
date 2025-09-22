/**
 * JavaScript code to add MJSET report generation buttons to Gophish campaign results page
 * This can be integrated into the existing Gophish UI templates
 */

// Add report generation section to campaign results page
function addReportSection(campaignId) {
    const reportSection = `
        <div class="panel panel-default" id="mjset-reports">
            <div class="panel-heading">
                <h3 class="panel-title">
                    <i class="fa fa-file-pdf-o"></i> MJSET Reports
                </h3>
            </div>
            <div class="panel-body">
                <p>Generate professional PDF reports for this campaign using MJSET formatting.</p>
                
                <div class="row">
                    <div class="col-md-4">
                        <button class="btn btn-primary btn-block" onclick="generateReport('${campaignId}', 'executive')">
                            <i class="fa fa-file-text"></i> Executive Summary
                        </button>
                        <small class="text-muted">High-level overview for management</small>
                    </div>
                    <div class="col-md-4">
                        <button class="btn btn-warning btn-block" onclick="generateReport('${campaignId}', 'confidential')">
                            <i class="fa fa-lock"></i> Confidential Report
                        </button>
                        <small class="text-muted">Detailed analysis with obfuscated data</small>
                    </div>
                    <div class="col-md-4">
                        <button class="btn btn-danger btn-block" onclick="generateReport('${campaignId}', 'internal')">
                            <i class="fa fa-eye"></i> Internal Report
                        </button>
                        <small class="text-muted">Complete data for security team</small>
                    </div>
                </div>
                
                <hr>
                
                <div class="row">
                    <div class="col-md-12">
                        <button class="btn btn-success btn-lg btn-block" onclick="generateAllReports('${campaignId}')">
                            <i class="fa fa-cogs"></i> Generate All Reports
                        </button>
                    </div>
                </div>
                
                <div id="report-status" class="alert" style="display: none; margin-top: 15px;"></div>
                
                <div id="report-downloads" style="margin-top: 15px; display: none;">
                    <h4>Download Reports:</h4>
                    <div class="list-group" id="download-links"></div>
                </div>
            </div>
        </div>
    `;
    
    // Insert after the campaign timeline section
    const timelineSection = document.querySelector('#timeline');
    if (timelineSection) {
        timelineSection.insertAdjacentHTML('afterend', reportSection);
    }
}

// Generate a specific report type
function generateReport(campaignId, reportType) {
    showStatus('info', `Generating ${reportType} report...`, true);
    
    fetch(`/api/campaigns/${campaignId}/reports/generate`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${api.api_key}`,
            'Content-Type': 'application/json'
        }
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            showStatus('success', 'Report generated successfully!');
            addDownloadLink(campaignId, reportType);
        } else {
            showStatus('danger', `Error: ${data.message}`);
        }
    })
    .catch(error => {
        showStatus('danger', `Error generating report: ${error.message}`);
    });
}

// Generate all report types
function generateAllReports(campaignId) {
    showStatus('info', 'Generating all reports...', true);
    
    const reportTypes = ['executive', 'confidential', 'internal'];
    let completed = 0;
    
    // Generate each report type
    reportTypes.forEach(reportType => {
        fetch(`/api/campaigns/${campaignId}/reports/generate`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${api.api_key}`,
                'Content-Type': 'application/json'
            }
        })
        .then(response => response.json())
        .then(data => {
            completed++;
            if (data.success) {
                addDownloadLink(campaignId, reportType);
            }
            
            if (completed === reportTypes.length) {
                showStatus('success', 'All reports generated successfully!');
            }
        })
        .catch(error => {
            completed++;
            console.error(`Error generating ${reportType} report:`, error);
            
            if (completed === reportTypes.length) {
                showStatus('warning', 'Some reports may have failed. Check console for details.');
            }
        });
    });
}

// Show status message
function showStatus(type, message, showSpinner = false) {
    const statusDiv = document.getElementById('report-status');
    statusDiv.className = `alert alert-${type}`;
    statusDiv.style.display = 'block';
    
    const spinner = showSpinner ? '<i class="fa fa-spinner fa-spin"></i> ' : '';
    statusDiv.innerHTML = spinner + message;
    
    if (!showSpinner) {
        // Hide status after 5 seconds
        setTimeout(() => {
            statusDiv.style.display = 'none';
        }, 5000);
    }
}

// Add download link for generated report
function addDownloadLink(campaignId, reportType) {
    const downloadsDiv = document.getElementById('report-downloads');
    const linksDiv = document.getElementById('download-links');
    
    downloadsDiv.style.display = 'block';
    
    const reportNames = {
        'executive': 'Executive Summary',
        'confidential': 'Confidential Report', 
        'internal': 'Internal Report'
    };
    
    const linkHtml = `
        <a href="/api/campaigns/${campaignId}/reports/download/${reportType}" 
           class="list-group-item" 
           target="_blank">
            <i class="fa fa-download"></i> ${reportNames[reportType]} 
            <span class="badge">PDF</span>
        </a>
    `;
    
    // Remove existing link for this report type
    const existingLink = linksDiv.querySelector(`a[href*="download/${reportType}"]`);
    if (existingLink) {
        existingLink.remove();
    }
    
    linksDiv.insertAdjacentHTML('beforeend', linkHtml);
}

// Initialize when page loads
document.addEventListener('DOMContentLoaded', function() {
    // Check if we're on a campaign results page
    const campaignIdMatch = window.location.pathname.match(/\/campaigns\/(\d+)/);
    if (campaignIdMatch) {
        const campaignId = campaignIdMatch[1];
        
        // Wait for the page to fully load, then add the report section
        setTimeout(() => {
            addReportSection(campaignId);
        }, 1000);
    }
});

// CSS styles to add to the page
const reportStyles = `
<style>
#mjset-reports .btn {
    margin-bottom: 5px;
}

#mjset-reports .text-muted {
    font-size: 11px;
    display: block;
    margin-top: 5px;
}

#mjset-reports .list-group-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

#mjset-reports .list-group-item:hover {
    background-color: #f5f5f5;
}

#mjset-reports .fa-spinner {
    margin-right: 5px;
}
</style>
`;

// Add styles to page
document.head.insertAdjacentHTML('beforeend', reportStyles);
