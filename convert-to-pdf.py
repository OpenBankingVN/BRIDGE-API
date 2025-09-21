#!/usr/bin/env python3

import webbrowser
import os
import sys

def convert_to_pdf():
    # Get the current directory
    current_dir = os.path.dirname(os.path.abspath(__file__))
    
    # Path to the HTML file
    html_file = os.path.join(current_dir, 'docs/architect/version_1/diagrams.html')
    
    # Check if file exists
    if not os.path.exists(html_file):
        print(f"❌ HTML file not found: {html_file}")
        return False
    
    # Convert to file URL
    file_url = f"file://{html_file}"
    
    print("🌐 Opening HTML file in browser for PDF conversion...")
    print(f"📄 File: {html_file}")
    print("\n📋 Instructions to convert to PDF:")
    print("1. Wait for the page to load completely (all Mermaid diagrams should render)")
    print("2. Press Cmd+P (on Mac) or Ctrl+P (on Windows/Linux)")
    print("3. Select 'Save as PDF' as destination")
    print("4. Click 'More settings' or 'Options'")
    print("5. Enable 'Background graphics' (important for diagram colors)")
    print("6. Set margins to 'Minimum' or 'None' for better layout")
    print("7. Click 'Save' and choose your PDF location")
    print("\n✅ The PDF will be generated with all your updated diagrams!")
    print("\n💡 Tip: Make sure to wait for all Mermaid diagrams to fully render before printing")
    
    # Open in default browser
    webbrowser.open(file_url)
    return True

if __name__ == "__main__":
    success = convert_to_pdf()
    sys.exit(0 if success else 1)
