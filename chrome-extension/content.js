/**
 * EduApps Chrome Extension - Content Script
 * Injected into web pages to detect text and provide practice features
 */

const API_BASE_URL = 'https://localhost:5051';

// Initialize content script
console.log('[EduApps] Content script loaded on:', window.location.hostname);

// Listen for messages from background script
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  console.log('[EduApps] Content script message:', request.action);

  switch (request.action) {
    case 'GET_SELECTED_TEXT':
      handleGetSelectedText(sendResponse);
      break;

    case 'INJECT_TOOLBAR':
      handleInjectToolbar(request, sendResponse);
      break;

    case 'ANALYZE_PAGE':
      handleAnalyzePage(sendResponse);
      break;

    default:
      sendResponse({ error: 'Unknown action' });
  }

  return true;
});

// Get selected text for typing practice
function handleGetSelectedText(sendResponse) {
  const selectedText = window.getSelection().toString();
  sendResponse({
    text: selectedText,
    length: selectedText.length
  });
}

// Inject floating toolbar into page
function handleInjectToolbar(request, sendResponse) {
  // Check if toolbar already exists
  if (document.getElementById('eduapps-toolbar')) {
    sendResponse({ error: 'Toolbar already injected' });
    return;
  }

  // Create toolbar container
  const toolbar = document.createElement('div');
  toolbar.id = 'eduapps-toolbar';
  toolbar.innerHTML = `
    <div style="
      position: fixed;
      bottom: 20px;
      right: 20px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      border-radius: 12px;
      padding: 12px;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
      z-index: 999999;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      color: white;
      user-select: none;
    ">
      <div style="display: flex; gap: 8px; margin-bottom: 8px;">
        <button class="eduapps-btn" data-action="typing" style="
          background: #ff6b6b;
          color: white;
          border: none;
          padding: 8px 12px;
          border-radius: 6px;
          cursor: pointer;
          font-size: 12px;
          font-weight: 500;
          transition: transform 0.2s;
        ">⌨ Typing</button>
        <button class="eduapps-btn" data-action="math" style="
          background: #4ecdc4;
          color: white;
          border: none;
          padding: 8px 12px;
          border-radius: 6px;
          cursor: pointer;
          font-size: 12px;
          font-weight: 500;
          transition: transform 0.2s;
        ">🔢 Math</button>
        <button class="eduapps-btn" data-action="reading" style="
          background: #ffd93d;
          color: #333;
          border: none;
          padding: 8px 12px;
          border-radius: 6px;
          cursor: pointer;
          font-size: 12px;
          font-weight: 500;
          transition: transform 0.2s;
        ">📖 Reading</button>
      </div>
      <div style="
        font-size: 11px;
        opacity: 0.9;
        text-align: center;
        cursor: pointer;
      " class="eduapps-minimize">▼ Hide</div>
    </div>
  `;

  document.body.appendChild(toolbar);

  // Add event listeners
  toolbar.querySelectorAll('.eduapps-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
      e.preventDefault();
      const action = btn.dataset.action;
      chrome.runtime.sendMessage({
        action: 'LAUNCH_PRACTICE',
        practiceType: action
      });
    });

    btn.addEventListener('mouseover', (e) => {
      e.target.style.transform = 'scale(1.05)';
    });

    btn.addEventListener('mouseout', (e) => {
      e.target.style.transform = 'scale(1)';
    });
  });

  // Add minimize functionality
  toolbar.querySelector('.eduapps-minimize').addEventListener('click', () => {
    const buttons = toolbar.querySelector('[style*="display: flex"]');
    if (buttons.style.display === 'none') {
      buttons.style.display = 'flex';
      toolbar.querySelector('.eduapps-minimize').textContent = '▼ Hide';
    } else {
      buttons.style.display = 'none';
      toolbar.querySelector('.eduapps-minimize').textContent = '▲ Show';
    }
  });

  sendResponse({ success: true, message: 'Toolbar injected' });
}

// Analyze page content for practice opportunities
function handleAnalyzePage(sendResponse) {
  const analysis = {
    title: document.title,
    textContent: document.body.innerText.substring(0, 500),
    images: document.querySelectorAll('img').length,
    links: document.querySelectorAll('a').length,
    headers: document.querySelectorAll('h1, h2, h3').length,
    practiceOpportunities: []
  };

  // Detect practice opportunities
  const text = document.body.innerText.toLowerCase();

  if (text.includes('math') || text.includes('calculate') || text.match(/\d+\s*[+\-*/]\s*\d+/)) {
    analysis.practiceOpportunities.push('math');
  }

  if (text.length > 500) {
    analysis.practiceOpportunities.push('reading');
  }

  // Always allow typing practice
  analysis.practiceOpportunities.push('typing');

  sendResponse(analysis);
}

// Detect when user selects text and show practice suggestion
document.addEventListener('mouseup', debounce(() => {
  const selectedText = window.getSelection().toString();

  if (selectedText.length > 10 && selectedText.length < 5000) {
    // Show subtle hint (optional)
    console.log('[EduApps] Selected text (length: ' + selectedText.length + '): ' + selectedText.substring(0, 50) + '...');
  }
}, 300));

// Utility function for debouncing
function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

// Handle keyboard shortcuts
document.addEventListener('keydown', (e) => {
  // Ctrl+Shift+E to toggle toolbar
  if (e.ctrlKey && e.shiftKey && e.code === 'KeyE') {
    chrome.runtime.sendMessage({
      action: 'TOGGLE_TOOLBAR'
    });
  }

  // Ctrl+Shift+T for quick typing
  if (e.ctrlKey && e.shiftKey && e.code === 'KeyT') {
    chrome.tabs.create({ url: `${API_BASE_URL}/typing` });
  }
});

// Inject injected.js for specific functionality
const script = document.createElement('script');
script.src = chrome.runtime.getURL('injected.js');
script.onload = function() {
  this.remove();
};
(document.head || document.documentElement).appendChild(script);

console.log('[EduApps] Content script initialized');
