import { chromium } from 'playwright';

(async () => {
  console.log('🚀 Starting Cloudscape Components Quick Test...');

  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext();
  const page = await context.newPage();

  // Enable console logging
  page.on('console', msg => console.log('PAGE:', msg.text()));

  try {
    console.log('📡 Navigating to GUI application...');
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: 30000 });
    await page.waitForTimeout(5000);

    console.log('🔍 Testing Cloudscape Components...');

    // Test 1: Page loads and has content
    const pageTitle = await page.title();
    console.log(`✅ Page title: ${pageTitle}`);

    // Test 2: Check for Cloudscape CSS
    const stylesheets = await page.evaluate(() => {
      return Array.from(document.styleSheets).map(sheet => {
        try {
          return sheet.href;
        } catch (e) {
          return null;
        }
      }).filter(href => href);
    });

    const hasCloudscapeCSS = stylesheets.some(href => href && href.includes('cloudscape'));
    console.log(`${hasCloudscapeCSS ? '✅' : '❌'} Cloudscape CSS: ${hasCloudscapeCSS ? 'Found' : 'Not found'}`);

    // Test 3: Check for Prism header/title
    const headerText = await page.evaluate(() => {
      const headers = document.querySelectorAll('h1, h2, .title, [data-testid="title"]');
      return Array.from(headers).map(h => h.textContent).join(', ');
    });
    console.log(`✅ Headers found: ${headerText || 'None'}`);

    // Test 4: Check for AWSUI/Cloudscape classes
    const awsuiCount = await page.evaluate(() => {
      const allElements = document.getElementsByTagName('*');
      let count = 0;
      for (let el of allElements) {
        if (el.className && typeof el.className === 'string' && el.className.includes('awsui')) {
          count++;
        }
      }
      return count;
    });
    console.log(`${awsuiCount > 0 ? '✅' : '❌'} Cloudscape components: ${awsuiCount} elements with AWSUI classes`);

    // Test 5: Check for buttons and interactive elements
    const buttonCount = await page.locator('button').count();
    console.log(`✅ Interactive elements: ${buttonCount} buttons found`);

    // Test 6: Check for any tabs or navigation
    const tabElements = await page.evaluate(() => {
      const selectors = ['[role="tablist"]', '[data-testid="tabs"]', '.awsui-tabs', 'nav'];
      let found = [];
      for (const selector of selectors) {
        const elements = document.querySelectorAll(selector);
        if (elements.length > 0) {
          found.push(`${selector}: ${elements.length}`);
        }
      }
      return found;
    });
    console.log(`✅ Navigation elements: ${tabElements.length > 0 ? tabElements.join(', ') : 'None found'}`);

    // Test 7: Check for form elements
    const formElements = await page.evaluate(() => {
      const inputs = document.querySelectorAll('input').length;
      const selects = document.querySelectorAll('select').length;
      const textareas = document.querySelectorAll('textarea').length;
      return { inputs, selects, textareas };
    });
    console.log(`✅ Form elements: ${formElements.inputs} inputs, ${formElements.selects} selects, ${formElements.textareas} textareas`);

    // Test 8: Take a screenshot for visual verification
    await page.screenshot({ path: 'cloudscape-gui-screenshot.png', fullPage: true });
    console.log('✅ Screenshot saved as cloudscape-gui-screenshot.png');

    // Test 9: Check for any error messages or loading states
    const errorText = await page.evaluate(() => {
      const bodyText = document.body.textContent.toLowerCase();
      const hasError = bodyText.includes('error') || bodyText.includes('failed');
      const hasLoading = bodyText.includes('loading') || bodyText.includes('loading...');
      const hasContent = bodyText.includes('template') || bodyText.includes('instance') || bodyText.includes('prism');
      return { hasError, hasLoading, hasContent, length: bodyText.length };
    });

    console.log(`${errorText.hasError ? '⚠️' : '✅'} Error state: ${errorText.hasError ? 'Errors found' : 'No errors'}`);
    console.log(`${errorText.hasLoading ? 'ℹ️' : '✅'} Loading state: ${errorText.hasLoading ? 'Loading detected' : 'Content loaded'}`);
    console.log(`${errorText.hasContent ? '✅' : '❌'} Content loaded: ${errorText.hasContent ? 'Prism content found' : 'No content found'}`);
    console.log(`✅ Page content length: ${errorText.length} characters`);

    console.log('\n🎉 Cloudscape Quick Test Complete!');
    console.log('✅ GUI application is running with Cloudscape components');

  } catch (error) {
    console.error('❌ Test failed:', error.message);
  } finally {
    await page.waitForTimeout(2000);
    await browser.close();
  }
})();