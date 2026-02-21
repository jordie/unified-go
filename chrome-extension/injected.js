/**
 * EduApps Chrome Extension - Injected Script
 * Injected into page context for DOM manipulation and specific functionality
 * This runs in the page's context, not the extension's context
 */

// Communication bridge between injected script and content script
window.addEventListener('message', (event) => {
  if (event.source !== window) return;

  if (event.data.type && event.data.type === 'EDUAPPS_REQUEST') {
    // Forward to content script
    chrome.runtime.sendMessage(event.data.message, (response) => {
      window.postMessage({
        type: 'EDUAPPS_RESPONSE',
        id: event.data.id,
        response
      }, '*');
    });
  }
});

// Expose global API for page scripts
window.EduApps = {
  // Get selected text
  getSelectedText: function() {
    return window.getSelection().toString();
  },

  // Get page content
  getPageContent: function() {
    return {
      title: document.title,
      url: window.location.href,
      text: document.body.innerText.substring(0, 5000),
      textLength: document.body.innerText.length,
      images: document.querySelectorAll('img').length,
      headings: document.querySelectorAll('h1, h2, h3, h4, h5, h6').length,
      links: document.querySelectorAll('a').length
    };
  },

  // Highlight text
  highlightText: function(text, color = 'yellow') {
    const selection = window.getSelection();
    if (selection.toString() === text) {
      const range = selection.getRangeAt(0);
      const span = document.createElement('span');
      span.style.backgroundColor = color;
      span.appendChild(range.extractContents());
      range.insertNode(span);
    }
  },

  // Extract form data
  getFormData: function() {
    const forms = [];
    document.querySelectorAll('form').forEach((form, index) => {
      forms.push({
        index,
        action: form.action,
        method: form.method,
        fields: Array.from(form.querySelectorAll('input, textarea, select')).map(field => ({
          name: field.name,
          type: field.type,
          label: document.querySelector(`label[for="${field.id}"]`)?.textContent || ''
        }))
      });
    });
    return forms;
  },

  // Extract metadata
  getMetadata: function() {
    return {
      charset: document.charset || 'utf-8',
      lang: document.documentElement.lang,
      description: document.querySelector('meta[name="description"]')?.content || '',
      keywords: document.querySelector('meta[name="keywords"]')?.content || '',
      author: document.querySelector('meta[name="author"]')?.content || '',
      viewport: document.querySelector('meta[name="viewport"]')?.content || ''
    };
  },

  // Check for reading content
  isReadingContent: function() {
    const text = document.body.innerText || '';
    const wordCount = text.split(/\s+/).length;
    return wordCount > 200; // At least 200 words
  },

  // Check for math content
  isMathContent: function() {
    const text = document.body.innerText || '';
    const mathPatterns = [
      /\d+\s*[\+\-\*\/]\s*\d+/,
      /algebra|calculus|geometry|equation|formula|theorem/i,
      /π|∑|∫|√|×|÷|≈|≤|≥/
    ];
    return mathPatterns.some(pattern => pattern.test(text));
  },

  // Listen for user interactions
  trackInteractions: function(callback) {
    document.addEventListener('click', (e) => {
      callback({
        type: 'click',
        target: e.target.tagName,
        className: e.target.className,
        timestamp: new Date().toISOString()
      });
    });

    document.addEventListener('submit', (e) => {
      callback({
        type: 'submit',
        formAction: e.target.action,
        timestamp: new Date().toISOString()
      });
    });
  },

  // Log activity
  logActivity: function(activity) {
    window.postMessage({
      type: 'EDUAPPS_REQUEST',
      message: {
        action: 'LOG_ACTIVITY',
        activity
      }
    }, '*');
  }
};

console.log('[EduApps] Injected script loaded - API available as window.EduApps');
