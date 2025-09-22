#!/usr/bin/env python3
"""
Test script for Gophish-MJSET integration
This script tests the report generation without requiring a running Gophish instance
"""

import json
import os
import sys
from datetime import datetime

# Add the mjset directory to the Python path
sys.path.append(os.path.join(os.path.dirname(__file__), 'mjset'))

from gophish_bridge import GophishBridge

def create_test_data():
    """Create sample campaign data for testing"""
    return {
        "campaign": {
            "id": 1,
            "name": "Test Campaign",
            "launch_date": "2024-01-15T10:00:00Z",
            "created_date": "2024-01-14T15:30:00Z",
            "url": "https://test-domain.com",
            "status": "Completed"
        },
        "credentials": [
            {
                "email": "user1@example.com",
                "username": "user1",
                "password": "password123",
                "ip": "192.168.1.100",
                "timestamp": "2024-01-15T10:15:00Z"
            },
            {
                "email": "user2@example.com", 
                "username": "user2",
                "password": "admin",
                "ip": "192.168.1.101",
                "timestamp": "2024-01-15T10:20:00Z"
            }
        ],
        "clicks": [
            {
                "email": "user1@example.com",
                "ip": "192.168.1.100",
                "timestamp": "2024-01-15T10:10:00Z"
            },
            {
                "email": "user2@example.com",
                "ip": "192.168.1.101", 
                "timestamp": "2024-01-15T10:12:00Z"
            },
            {
                "email": "user3@example.com",
                "ip": "192.168.1.102",
                "timestamp": "2024-01-15T10:25:00Z"
            }
        ],
        "targets": [
            "user1@example.com",
            "user2@example.com", 
            "user3@example.com",
            "user4@example.com",
            "user5@example.com"
        ],
        "stats": {
            "total_targets": 5,
            "emails_sent": 5,
            "emails_opened": 4,
            "links_clicked": 3,
            "cred_submitted": 2,
            "emails_reported": 0
        }
    }

def main():
    print("Testing Gophish-MJSET Integration...")
    
    # Create test data
    test_data = create_test_data()
    
    # Create a mock bridge instance
    bridge = GophishBridge("test")
    
    try:
        # Test report generation with mock data
        print("Generating test reports...")
        output_dir = bridge.generate_reports(test_data)
        
        # Check if reports were created
        expected_files = ["SocialEngExecReport.pdf"]
        
        for filename in expected_files:
            filepath = os.path.join(output_dir, filename)
            if os.path.exists(filepath):
                size = os.path.getsize(filepath)
                print(f"✓ {filename} created successfully ({size} bytes)")
            else:
                print(f"✗ {filename} not found")
        
        print(f"\nTest completed! Reports saved to: {output_dir}")
        
    except Exception as e:
        print(f"Test failed: {e}")
        import traceback
        traceback.print_exc()
        return 1
    
    return 0

if __name__ == "__main__":
    sys.exit(main())
