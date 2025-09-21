const puppeteer = require('puppeteer');
const path = require('path');

async function generatePDF() {
    const browser = await puppeteer.launch({
        headless: true,
        args: ['--no-sandbox', '--disable-setuid-sandbox']
    });
    
    const page = await browser.newPage();
    
    // Set viewport for better rendering
    await page.setViewport({
        width: 1200,
        height: 800,
        deviceScaleFactor: 2
    });
    
    // Load the HTML file
    const htmlPath = path.resolve(__dirname, 'docs/architect/version_1/diagrams.html');
    await page.goto(`file://${htmlPath}`, {
        waitUntil: 'networkidle0'
    });
    
    // Wait for Mermaid diagrams to render
    await page.waitForTimeout(3000);
    
    // Generate PDF
    const pdfPath = path.resolve(__dirname, 'docs/architect/version_1/BRIDGE-API-Diagrams.pdf');
    await page.pdf({
        path: pdfPath,
        format: 'A4',
        printBackground: true,
        margin: {
            top: '20mm',
            right: '20mm',
            bottom: '20mm',
            left: '20mm'
        },
        displayHeaderFooter: true,
        headerTemplate: '<div style="font-size: 10px; text-align: center; width: 100%; color: #666;">BRIDGE-API Flow Diagrams</div>',
        footerTemplate: '<div style="font-size: 10px; text-align: center; width: 100%; color: #666;">Page <span class="pageNumber"></span> of <span class="totalPages"></span></div>'
    });
    
    await browser.close();
    
    console.log(`✅ PDF generated successfully: ${pdfPath}`);
}

generatePDF().catch(console.error);
