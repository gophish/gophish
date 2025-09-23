#!/usr/bin/env python3
"""
Gophish to MJSET Bridge Script
This script fetches campaign data from Gophish API and generates MJSET-style reports.
"""

import sys
import os
import json
import requests
import tempfile
import shutil
from datetime import datetime
from pathlib import Path

# Add the mjset directory to the Python path
sys.path.append(os.path.join(os.path.dirname(__file__), 'mjset'))

from mjset.PdfReport import PdfReport
from mjset.makeSEReport import makeSEReport


class GophishBridge:
    def __init__(self, campaign_id, gophish_url="https://localhost:3333", api_key=None):
        self.campaign_id = campaign_id
        self.gophish_url = gophish_url.rstrip('/')
        self.api_key = api_key or self._get_api_key()
        self.temp_dir = None
        self.output_dir = None
        
    def _get_api_key(self):
        """Get API key from environment or config file"""
        # Try environment variable first
        api_key = os.environ.get('GOPHISH_API_KEY')
        if api_key:
            return api_key
            
        # Try to read from config file
        config_path = os.path.join(os.path.dirname(__file__), '..', 'config.json')
        if os.path.exists(config_path):
            try:
                with open(config_path, 'r') as f:
                    config = json.load(f)
                    return config.get('admin_server', {}).get('api_key', '')
            except:
                pass
                
        # Default API key for development (should be changed in production)
        return "gophish"
    
    def fetch_campaign_data(self):
        """Fetch campaign data from Gophish API"""
        headers = {
            'Authorization': f'Bearer {self.api_key}',
            'Content-Type': 'application/json'
        }
        
        try:
            response = requests.get(
                f'{self.gophish_url}/api/campaigns/{self.campaign_id}/export',
                headers=headers,
                verify=False  # For self-signed certificates
            )
            response.raise_for_status()
            return response.json()
        except requests.exceptions.RequestException as e:
            print(f"Error fetching campaign data: {e}")
            sys.exit(1)
    
    def create_temp_files(self, data):
        """Create temporary files in MJSET format"""
        self.temp_dir = tempfile.mkdtemp(prefix='gophish_reports_')
        
        # Create cred.csv file
        cred_file = os.path.join(self.temp_dir, 'cred.csv')
        with open(cred_file, 'w') as f:
            for cred in data['credentials']:
                # Format: user|username|password|sourceIP|timestamp
                timestamp = datetime.fromisoformat(cred['timestamp'].replace('Z', '+00:00')).strftime('%Y-%m-%d %H:%M:%S')
                f.write(f"{cred['email']}|{cred['username']}|{cred['password']}|{cred['ip']}|{timestamp}\n")
        
        # Create hits_report.txt file
        hits_file = os.path.join(self.temp_dir, 'hits_report.txt')
        click_counts = {}
        cred_counts = {}
        
        # Count clicks per email
        for click in data['clicks']:
            email = click['email']
            click_counts[email] = click_counts.get(email, 0) + 1
        
        # Count credentials per email
        for cred in data['credentials']:
            email = cred['email']
            cred_counts[email] = cred_counts.get(email, 0) + 1
        
        with open(hits_file, 'w') as f:
            # Include all targets, even those who didn't click
            for email in data['targets']:
                clicks = click_counts.get(email, 0)
                creds = cred_counts.get(email, 0)
                f.write(f"{email}|{clicks}|{creds}\n")
        
        # Create emails_mapped.txt file
        emails_file = os.path.join(self.temp_dir, 'emails_mapped.txt')
        with open(emails_file, 'w') as f:
            for email in data['targets']:
                f.write(f"{email}|\n")
        
        # Create scandata.txt file
        scandata_file = os.path.join(self.temp_dir, 'scandata.txt')
        campaign = data['campaign']
        with open(scandata_file, 'w') as f:
            f.write(f"Client Name = {campaign['name']}\n")
            launch_date = datetime.fromisoformat(campaign['launch_date'].replace('Z', '+00:00')).strftime('%Y-%m-%d')
            f.write(f"Scan Started = {launch_date}\n")
            f.write(f"Malicious URLs = {campaign['url']}\n")
            f.write(f"Impersonated = Unknown\n")  # This would need to be configured
        
        return self.temp_dir
    
    def setup_output_directory(self):
        """Create output directory for reports"""
        reports_dir = os.path.join(os.path.dirname(__file__), 'output')
        self.output_dir = os.path.join(reports_dir, f'campaign_{self.campaign_id}')
        os.makedirs(self.output_dir, exist_ok=True)
        return self.output_dir
    
    def generate_reports(self, data):
        """Generate MJSET-style reports"""
        temp_dir = self.create_temp_files(data)
        output_dir = self.setup_output_directory()
        
        try:
            # Create the MJSET report generator
            report_generator = makeSEReport()
            
            # Set up the report generator with our data
            report_generator.setClientFolder(temp_dir + '/')
            report_generator.setDomain(data['campaign']['url'])
            
            # Override the output path to use our output directory
            report_generator.companyReportPath = output_dir + '/'
            
            # Set client name from campaign
            report_generator.setClientName(data['campaign']['name'])
            
            # Initialize data (this normally prompts for user input, so we'll mock it)
            self._mock_user_input(report_generator, data)
            
            # Generate the reports
            report_generator.writeReport()
            
            print(f"Reports generated successfully in: {output_dir}")
            return output_dir
            
        except Exception as e:
            print(f"Error generating reports: {e}")
            raise
        finally:
            # Clean up temporary files
            if temp_dir and os.path.exists(temp_dir):
                shutil.rmtree(temp_dir)
    
    def _mock_user_input(self, report_generator, data):
        """Mock the user input that MJSET normally requires"""
        # Set default values that would normally be prompted
        report_generator.cA1 = "Client Address Line 1"
        report_generator.cA2 = "Client Address Line 2"
        report_generator.initialBlocked = 'n'  # Assume not initially blocked
        report_generator.testMethod = "externally using email links"
        report_generator.navVecString = 'clicking on a link'
        report_generator.blockedString = 'were not initially'
        report_generator.impersonated = 'Unknown Organization'
        
        # Set up file paths
        report_generator.credCSVFile = os.path.join(report_generator.clientFolder, 'cred.csv')
        report_generator.emailsMappedTxt = os.path.join(report_generator.clientFolder, 'emails_mapped.txt')
        report_generator.hitsReportTextFile = os.path.join(report_generator.clientFolder, 'hits_report.txt')
        report_generator.scanDataFile = os.path.join(report_generator.clientFolder, 'scandata.txt')
        
        # Create dummy screenshot files (these would normally be provided by user)
        dummy_screenshot = os.path.join(report_generator.clientFolder, 'screenshot.png')
        dummy_email = os.path.join(report_generator.clientFolder, 'email.png')
        
        # Create minimal 1x1 pixel PNG files as placeholders
        self._create_dummy_image(dummy_screenshot)
        self._create_dummy_image(dummy_email)
        
        report_generator.screenshotFile = dummy_screenshot
        report_generator.emailScreenshot = dummy_email
        
        # Process the data as MJSET would
        report_generator._process_mjset_data()
    
    def _create_dummy_image(self, filepath):
        """Create a minimal PNG file as placeholder"""
        # Minimal 1x1 pixel PNG file content
        png_data = b'\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\tpHYs\x00\x00\x0b\x13\x00\x00\x0b\x13\x01\x00\x9a\x9c\x18\x00\x00\x00\nIDATx\x9cc\xf8\x00\x00\x00\x01\x00\x01\x00\x00\x00\x00IEND\xaeB`\x82'
        with open(filepath, 'wb') as f:
            f.write(png_data)


def main():
    if len(sys.argv) != 2:
        print("Usage: python3 gophish_bridge.py <campaign_id>")
        sys.exit(1)
    
    campaign_id = sys.argv[1]
    
    try:
        # Create bridge instance
        bridge = GophishBridge(campaign_id)
        
        # Fetch campaign data from Gophish
        print(f"Fetching campaign data for campaign {campaign_id}...")
        data = bridge.fetch_campaign_data()
        
        # Generate reports
        print("Generating reports...")
        output_dir = bridge.generate_reports(data)
        
        print(f"Report generation completed successfully!")
        print(f"Reports saved to: {output_dir}")
        
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
