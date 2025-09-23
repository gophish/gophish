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
        
        # Use user customizations if available, otherwise use defaults
        if hasattr(self, 'user_customizations'):
            customizations = self.user_customizations
            report_generator.testMethod = customizations['test_method']
            report_generator.initialBlocked = customizations['blocked']
            report_generator.impersonated = customizations['impersonated']
            report_generator.blockedString = 'were initially' if customizations['blocked'] == 'y' else 'were not initially'
        else:
            # Default values
            report_generator.initialBlocked = 'n'
            report_generator.testMethod = "externally using email links"
            report_generator.blockedString = 'were not initially'
            
            # Try to extract sender info from campaign data or use default
            campaign_url = data['campaign'].get('url', 'Unknown Domain')
            if 'training.' in campaign_url:
                report_generator.impersonated = 'IT Training Department'
            elif 'bank' in campaign_url.lower():
                report_generator.impersonated = 'Banking Institution'
            else:
                report_generator.impersonated = 'IT Department'
        
        # Set navigation vector based on test method
        if 'QR' in report_generator.testMethod:
            report_generator.navVecString = 'scanning QR codes'
        else:
            report_generator.navVecString = 'clicking on a link'
        
        # Set up file paths
        report_generator.credCSVFile = os.path.join(report_generator.clientFolder, 'cred.csv')
        report_generator.emailsMappedTxt = os.path.join(report_generator.clientFolder, 'emails_mapped.txt')
        report_generator.hitsReportTextFile = os.path.join(report_generator.clientFolder, 'hits_report.txt')
        report_generator.scanDataFile = os.path.join(report_generator.clientFolder, 'scandata.txt')
        
        # Handle screenshots
        self._handle_screenshots(report_generator, data)
        
        # Process the data as MJSET would
        report_generator._process_mjset_data()
    
    def _handle_screenshots(self, report_generator, data):
        """Handle screenshot acquisition - either automatic or manual"""
        screenshot_dir = report_generator.clientFolder
        email_screenshot = os.path.join(screenshot_dir, 'email.png')
        website_screenshot = os.path.join(screenshot_dir, 'website.png')
        
        # First, try to get screenshots automatically from Gophish data
        auto_screenshots = self._try_automatic_screenshots(data, screenshot_dir)
        
        # If in interactive mode and auto screenshots not found, prompt user
        if hasattr(self, 'interactive_mode') and self.interactive_mode:
            if not auto_screenshots['email']:
                email_screenshot = self._prompt_for_screenshot('email', screenshot_dir)
            else:
                email_screenshot = auto_screenshots['email']
                
            if not auto_screenshots['website']:
                website_screenshot = self._prompt_for_screenshot('landing page', screenshot_dir)
            else:
                website_screenshot = auto_screenshots['website']
        else:
            # Non-interactive: use auto screenshots or create placeholders
            if auto_screenshots['email']:
                email_screenshot = auto_screenshots['email']
            else:
                self._create_dummy_image(email_screenshot)
                print("Note: No email screenshot found. Using placeholder.")
                
            if auto_screenshots['website']:
                website_screenshot = auto_screenshots['website']
            else:
                self._create_dummy_image(website_screenshot)
                print("Note: No website screenshot found. Using placeholder.")
        
        report_generator.emailScreenshot = email_screenshot
        report_generator.screenshotFile = website_screenshot
    
    def _try_automatic_screenshots(self, data, screenshot_dir):
        """Try to get screenshots from Gophish campaign data"""
        screenshots = {'email': None, 'website': None}
        
        # Check if campaign has template with HTML content
        try:
            # This would need to be implemented based on how Gophish stores templates
            # For now, check common screenshot locations
            common_paths = [
                ('email', ['email_screenshot.png', 'email.png', 'template.png']),
                ('website', ['website_screenshot.png', 'landing.png', 'page.png'])
            ]
            
            for screenshot_type, filenames in common_paths:
                for filename in filenames:
                    filepath = os.path.join(screenshot_dir, filename)
                    if os.path.exists(filepath):
                        screenshots[screenshot_type] = filepath
                        print(f"Found existing {screenshot_type} screenshot: {filename}")
                        break
        except Exception as e:
            print(f"Could not auto-detect screenshots: {e}")
        
        return screenshots
    
    def _prompt_for_screenshot(self, screenshot_type, screenshot_dir):
        """Prompt user to provide a screenshot file"""
        print(f"\n{'-'*60}")
        print(f"SCREENSHOT REQUIRED: {screenshot_type.upper()}")
        print(f"{'-'*60}")
        
        while True:
            print(f"\nPlease provide the {screenshot_type} screenshot:")
            print("Options:")
            print("1. Enter full path to existing screenshot file")
            print("2. Copy screenshot to: " + screenshot_dir)
            print("3. Skip (use placeholder)")
            
            choice = input("\nEnter your choice (1-3): ").strip()
            
            if choice == '1':
                filepath = input("Enter full path to screenshot: ").strip()
                if os.path.exists(filepath):
                    # Copy the file to our directory
                    import shutil
                    filename = f"{screenshot_type.replace(' ', '_')}_screenshot.png"
                    dest_path = os.path.join(screenshot_dir, filename)
                    shutil.copy2(filepath, dest_path)
                    print(f"Screenshot copied successfully!")
                    return dest_path
                else:
                    print("File not found. Please try again.")
                    
            elif choice == '2':
                filename = f"{screenshot_type.replace(' ', '_')}_screenshot.png"
                expected_path = os.path.join(screenshot_dir, filename)
                print(f"\nPlease copy your screenshot to:")
                print(f"  {expected_path}")
                input("\nPress Enter when done...")
                
                if os.path.exists(expected_path):
                    print("Screenshot found!")
                    return expected_path
                else:
                    print("File not found at expected location.")
                    
            elif choice == '3':
                # Create placeholder
                placeholder_path = os.path.join(screenshot_dir, f"{screenshot_type}_placeholder.png")
                self._create_dummy_image(placeholder_path)
                print(f"Using placeholder for {screenshot_type} screenshot.")
                return placeholder_path
            
            else:
                print("Invalid choice. Please enter 1, 2, or 3.")
    
    def _create_dummy_image(self, filepath):
        """Create a minimal PNG file as placeholder"""
        # Minimal 1x1 pixel PNG file content
        png_data = b'\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\tpHYs\x00\x00\x0b\x13\x00\x00\x0b\x13\x01\x00\x9a\x9c\x18\x00\x00\x00\nIDATx\x9cc\xf8\x00\x00\x00\x01\x00\x01\x00\x00\x00\x00IEND\xaeB`\x82'
        with open(filepath, 'wb') as f:
            f.write(png_data)


def get_user_input_for_report():
    """Get user input for report customization"""
    print("\n" + "="*60)
    print("REPORT CUSTOMIZATION")
    print("="*60)
    
    # Get test method
    print("\nComplete the sentence 'The test was performed ' with one of the following options:")
    print("     1) externally using email links")
    print("     2) externally using QR Codes") 
    print("     3) externally using email links and QR Codes")
    print("     4) internally using email links")
    print("     5) internally using QR Codes")
    print("     6) internally using email links and QR Codes")
    
    while True:
        choice = input("\nEnter choice (1-6): ").strip()
        test_methods = {
            '1': 'externally using email links',
            '2': 'externally using QR Codes',
            '3': 'externally using email links and QR Codes', 
            '4': 'internally using email links',
            '5': 'internally using QR Codes',
            '6': 'internally using email links and QR Codes'
        }
        if choice in test_methods:
            test_method = test_methods[choice]
            break
        print("Invalid choice. Please enter 1-6.")
    
    # Get blocking status
    while True:
        blocked = input("\nWas the attack initially blocked by client defenses? (y/n): ").strip().lower()
        if blocked in ['y', 'n']:
            break
        print("Please enter 'y' or 'n'")
    
    # Get impersonated entity
    impersonated = input("\nWho did you impersonate? (e.g., 'IT Department', 'HR Team'): ").strip()
    if not impersonated:
        impersonated = 'IT Department'
    
    return {
        'test_method': test_method,
        'blocked': blocked,
        'impersonated': impersonated
    }

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 gophish_bridge.py <campaign_id> [--interactive]")
        sys.exit(1)
    
    campaign_id = sys.argv[1]
    interactive = '--interactive' in sys.argv
    
    try:
        # Create bridge instance
        bridge = GophishBridge(campaign_id)
        bridge.interactive_mode = interactive
        
        # Fetch campaign data from Gophish
        print(f"Fetching campaign data for campaign {campaign_id}...")
        data = bridge.fetch_campaign_data()
        
        # Get user customizations if interactive mode
        if interactive:
            user_input = get_user_input_for_report()
            # Store user input for use in report generation
            bridge.user_customizations = user_input
        
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
