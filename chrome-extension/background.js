/**
 * EduApps Chrome Extension - Background Service Worker
 * Handles extension lifecycle, messages, notifications, and background tasks
 */

const API_BASE_URL = 'https://localhost:5051';
const EXTENSION_VERSION = '1.0.0';

// Initialize extension on install
chrome.runtime.onInstalled.addListener((details) => {
  if (details.reason === 'install') {
    // Open onboarding page
    chrome.tabs.create({
      url: chrome.runtime.getURL('onboarding.html')
    });

    // Initialize storage with default values
    chrome.storage.sync.set({
      extensionEnabled: true,
      userId: null,
      apiToken: null,
      syncInterval: 5, // minutes
      notificationsEnabled: true,
      practiceGoal: 30, // minutes per day
      lastSync: null,
      stats: {
        typingSessions: 0,
        mathSessions: 0,
        readingSessions: 0,
        totalMinutes: 0
      }
    });

    console.log('[EduApps] Extension installed and initialized');
  } else if (details.reason === 'update') {
    console.log(`[EduApps] Extension updated to version ${EXTENSION_VERSION}`);
  }
});

// Listen for messages from content scripts and popup
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  console.log('[EduApps] Message received:', request.action);

  switch (request.action) {
    case 'GET_USER':
      handleGetUser(sendResponse);
      break;

    case 'SYNC_DATA':
      handleSyncData(request, sendResponse);
      break;

    case 'LOG_SESSION':
      handleLogSession(request, sendResponse);
      break;

    case 'GET_STATS':
      handleGetStats(sendResponse);
      break;

    case 'SHOW_NOTIFICATION':
      handleShowNotification(request);
      break;

    case 'SET_USER':
      handleSetUser(request, sendResponse);
      break;

    case 'LOGOUT':
      handleLogout(sendResponse);
      break;

    default:
      sendResponse({ error: 'Unknown action' });
  }

  // Return true to indicate we'll send response asynchronously
  return true;
});

// Handle user login
async function handleSetUser(request, sendResponse) {
  const { userId, apiToken, username } = request;

  chrome.storage.sync.set({
    userId,
    apiToken,
    username,
    lastLogin: new Date().toISOString()
  }, () => {
    sendResponse({ success: true, message: 'User logged in' });
  });
}

// Handle user logout
async function handleLogout(sendResponse) {
  chrome.storage.sync.set({
    userId: null,
    apiToken: null,
    username: null
  }, () => {
    sendResponse({ success: true, message: 'Logged out' });
  });
}

// Fetch current user
function handleGetUser(sendResponse) {
  chrome.storage.sync.get(['userId', 'username', 'apiToken'], (data) => {
    sendResponse({
      userId: data.userId,
      username: data.username,
      isAuthenticated: !!data.userId
    });
  });
}

// Sync stats with server
async function handleSyncData(request, sendResponse) {
  try {
    const { userId, apiToken } = await chrome.storage.sync.get(['userId', 'apiToken']);

    if (!userId || !apiToken) {
      sendResponse({ error: 'Not authenticated' });
      return;
    }

    // Fetch latest stats from API
    const response = await fetch(`${API_BASE_URL}/api/dashboard/stats?user_id=${userId}`, {
      headers: {
        'Authorization': `Bearer ${apiToken}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      sendResponse({ error: `API error: ${response.status}` });
      return;
    }

    const data = await response.json();

    // Update local storage with synced data
    chrome.storage.sync.set({
      stats: data,
      lastSync: new Date().toISOString()
    });

    sendResponse({ success: true, data });
  } catch (error) {
    console.error('[EduApps] Sync error:', error);
    sendResponse({ error: error.message });
  }
}

// Log a practice session
async function handleLogSession(request, sendResponse) {
  try {
    const { app, duration, score, correctCount, totalCount } = request;
    const { userId, apiToken } = await chrome.storage.sync.get(['userId', 'apiToken']);

    if (!userId || !apiToken) {
      sendResponse({ error: 'Not authenticated' });
      return;
    }

    // Log session to API
    const response = await fetch(`${API_BASE_URL}/api/dashboard/analyze`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${apiToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        userId,
        app,
        duration,
        score,
        correctCount,
        totalCount,
        timestamp: new Date().toISOString()
      })
    });

    if (!response.ok) {
      sendResponse({ error: `Failed to log session: ${response.status}` });
      return;
    }

    // Update local stats
    chrome.storage.sync.get(['stats'], (data) => {
      const stats = data.stats || {};
      stats[`${app}Sessions`] = (stats[`${app}Sessions`] || 0) + 1;
      stats.totalMinutes = (stats.totalMinutes || 0) + Math.round(duration / 60);

      chrome.storage.sync.set({ stats }, () => {
        sendResponse({ success: true });
      });
    });
  } catch (error) {
    console.error('[EduApps] Session logging error:', error);
    sendResponse({ error: error.message });
  }
}

// Get stats from storage
function handleGetStats(sendResponse) {
  chrome.storage.sync.get(['stats', 'lastSync'], (data) => {
    sendResponse({
      stats: data.stats || {},
      lastSync: data.lastSync
    });
  });
}

// Show notification
function handleShowNotification(request) {
  const { title, message, iconUrl, clickUrl } = request;

  chrome.notifications.create({
    type: 'basic',
    title: title || 'EduApps',
    message: message || '',
    iconUrl: iconUrl || chrome.runtime.getURL('icons/icon-128.png'),
    priority: 1
  }, (notificationId) => {
    if (clickUrl) {
      chrome.notifications.onClicked.addListener((id) => {
        if (id === notificationId) {
          chrome.tabs.create({ url: clickUrl });
        }
      });
    }
  });
}

// Handle command shortcuts
chrome.commands.onCommand.addListener((command) => {
  console.log('[EduApps] Command executed:', command);

  switch (command) {
    case 'toggle-extension':
      chrome.storage.sync.get(['extensionEnabled'], (data) => {
        const enabled = !data.extensionEnabled;
        chrome.storage.sync.set({ extensionEnabled: enabled });
        console.log('[EduApps] Extension toggled:', enabled);
      });
      break;

    case 'quick-typing':
      chrome.tabs.create({ url: `${API_BASE_URL}/typing` });
      break;

    case 'quick-math':
      chrome.tabs.create({ url: `${API_BASE_URL}/math` });
      break;
  }
});

// Periodic sync (every N minutes)
function schedulePeriodicSync() {
  chrome.storage.sync.get(['syncInterval'], (data) => {
    const interval = (data.syncInterval || 5) * 60 * 1000; // convert to milliseconds

    setInterval(() => {
      chrome.storage.sync.get(['userId', 'apiToken'], (userData) => {
        if (userData.userId && userData.apiToken) {
          chrome.runtime.sendMessage({
            action: 'SYNC_DATA'
          }, (response) => {
            if (response && response.success) {
              console.log('[EduApps] Periodic sync completed');
            }
          });
        }
      });
    }, interval);
  });
}

// Start periodic sync on service worker startup
schedulePeriodicSync();

console.log('[EduApps] Background service worker loaded');
