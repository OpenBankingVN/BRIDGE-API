#!/usr/bin/env python3

import os
import sys
from weasyprint import HTML, CSS
from weasyprint.text.fonts import FontConfiguration

def generate_pdf():
    # Get the current directory
    current_dir = os.path.dirname(os.path.abspath(__file__))
    
    # Paths
    html_file = os.path.join(current_dir, 'docs/architect/version_1/diagrams.html')
    pdf_file = os.path.join(current_dir, 'docs/architect/version_1/BRIDGE-API-Diagrams.pdf')
    
    # Check if HTML file exists
    if not os.path.exists(html_file):
        print(f"❌ HTML file not found: {html_file}")
        return False
    
    try:
        # Create font configuration
        font_config = FontConfiguration()
        
        # Additional CSS for better PDF rendering
        additional_css = CSS(string='''
            @page {
                size: A4;
                margin: 20mm;
                @top-center {
                    content: "BRIDGE-API Flow Diagrams";
                    font-size: 10px;
                    color: #666;
                }
                @bottom-center {
                    content: "Page " counter(page) " of " counter(pages);
                    font-size: 10px;
                    color: #666;
                }
            }
            
            body {
                font-family: Arial, sans-serif;
                line-height: 1.4;
            }
            
            .container {
                max-width: none;
                margin: 0;
                padding: 0;
            }
            
            .diagram {
                page-break-inside: avoid;
                margin: 20px 0;
            }
            
            h1 {
                page-break-after: avoid;
            }
            
            h2 {
                page-break-after: avoid;
                page-break-inside: avoid;
            }
            
            .mermaid {
                text-align: center;
                page-break-inside: avoid;
            }
            
            .description {
                page-break-after: avoid;
            }
        ''', font_config=font_config)
        
        # Convert HTML to PDF
        print("🔄 Converting HTML to PDF...")
        HTML(filename=html_file).write_pdf(
            pdf_file,
            stylesheets=[additional_css],
            font_config=font_config
        )
        
        print(f"✅ PDF generated successfully: {pdf_file}")
        return True
        
    except Exception as e:
        print(f"❌ Error generating PDF: {str(e)}")
        return False

if __name__ == "__main__":
    success = generate_pdf()
    sys.exit(0 if success else 1)
