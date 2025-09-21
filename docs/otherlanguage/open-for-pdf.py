#!/usr/bin/env python3

import webbrowser
import os
import sys

def open_for_pdf():
    # Get the current directory
    current_dir = os.path.dirname(os.path.abspath(__file__))
    
    # Path to the print-optimized HTML file
    html_file = os.path.join(current_dir, 'docs/architect/version_1/diagrams-print.html')
    
    # Check if file exists
    if not os.path.exists(html_file):
        print(f"❌ HTML file not found: {html_file}")
        return False
    
    # Convert to file URL
    file_url = f"file://{html_file}"
    
    print("🌐 Opening HTML file in browser for PDF generation...")
    print(f"📄 File: {html_file}")
    print("\n📋 Instructions:")
    print("1. Wait for the page to load completely")
    print("2. Press Ctrl+P (Cmd+P on Mac)")
    print("3. Select 'Save as PDF' as destination")
    print("4. Click 'More settings'")
    print("5. Enable 'Background graphics'")
    print("6. Click 'Save'")
    print("\n✅ The PDF will be generated with all diagrams!")
    
    # Open in default browser
    webbrowser.open(file_url)
    return True

if __name__ == "__main__":
    success = open_for_pdf()
    sys.exit(0 if success else 1)
