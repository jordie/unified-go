/**
 * EduApps Chrome Extension - Options Script
 * Handles extension settings and preferences
 */

const DEFAULT_SETTINGS = {
  notificationsEnabled: true,
  notifyAchievements: true,
  notifyReminders: true,
  notifyMilestones: true,
  syncInterval: 5,
  autoSync: true,
  floatingToolbar: true,
  theme: 'auto',
  apiUrl: 'https://localhost:5051',
  debugMode: false
};

// Load settings on page load
document.addEventListener('DOMContentLoaded', () => {
  loadUserProfile();
  loadSettings();
  attachEventListeners();
});

// Load user profile
function loadUserProfile() {
  chrome.storage.sync.get(['userId', 'username'], (data) => {
    const profileSection = document.getElementById('profile-section');
    const loginSection = document.getElementById('login-section');

    if (data.userId && data.username) {
      profileSection.style.display = 'block';
      loginSection.style.display = 'none';

      document.getElementById('profile-username').textContent = data.username;
      document.getElementById('profile-avatar').textContent = data.username.charAt(0).toUpperCase();
    } else {
      profileSection.style.display = 'none';
      loginSection.style.display = 'block';
    }
  });
}

// Load all settings from storage
function loadSettings() {
  chrome.storage.sync.get(DEFAULT_SETTINGS, (data) => {
    // Notification settings
    document.getElementById('notifications-enabled').checked = data.notificationsEnabled;
    document.getElementById('notify-achievements').checked = data.notifyAchievements;
    document.getElementById('notify-reminders').checked = data.notifyReminders;
    document.getElementById('notify-milestones').checked = data.notifyMilestones;

    // Sync settings
    document.getElementById('sync-interval').value = data.syncInterval;
    document.getElementById('auto-sync').checked = data.autoSync;

    // Appearance settings
    document.getElementById('floating-toolbar').checked = data.floatingToolbar;
    document.getElementById('theme').value = data.theme;

    // API settings
    document.getElementById('api-url').value = data.apiUrl;

    // Advanced settings
    document.getElementById('debug-mode').checked = data.debugMode;
  });
}

// Attach event listeners
function attachEventListeners() {
  // Login
  document.getElementById('login-btn').addEventListener('click', handleLogin);
  document.getElementById('username').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') handleLogin();
  });
  document.getElementById('password').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') handleLogin();
  });

  // Logout
  document.getElementById('logout-btn')?.addEventListener('click', handleLogout);

  // Save button
  document.getElementById('save-btn').addEventListener('click', saveSettings);

  // Cancel button
  document.getElementById('cancel-btn').addEventListener('click', () => {
    loadSettings();
    showMessage('Changes discarded', 'info');
  });

  // Advanced buttons
  document.getElementById('clear-cache-btn').addEventListener('click', clearCache);
  document.getElementById('reset-settings-btn').addEventListener('click', resetToDefaults);

  // Notification type enable/disable
  document.getElementById('notifications-enabled').addEventListener('change', (e) => {
    const notifyCheckboxes = document.querySelectorAll('[id^="notify-"]');
    notifyCheckboxes.forEach(checkbox => {
      checkbox.disabled = !e.target.checked;
    });
  });
}

// Handle login
async function handleLogin() {
  const username = document.getElementById('username').value.trim();
  const password = document.getElementById('password').value.trim();

  if (!username || !password) {
    showMessage('Please enter username and password', 'error');
    return;
  }

  const loginBtn = document.getElementById('login-btn');
  loginBtn.disabled = true;
  loginBtn.textContent = 'Logging in...';

  try {
    const apiUrl = document.getElementById('api-url').value || DEFAULT_SETTINGS.apiUrl;

    const response = await fetch(`${apiUrl}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ username, password })
    });

    if (!response.ok) {
      throw new Error('Login failed');
    }

    const data = await response.json();

    // Store user data
    chrome.storage.sync.set({
      userId: data.userId,
      username: data.username,
      apiToken: data.token,
      lastLogin: new Date().toISOString()
    }, () => {
      document.getElementById('username').value = '';
      document.getElementById('password').value = '';
      loadUserProfile();
      showMessage('Login successful!', 'success');
    });
  } catch (error) {
    showMessage('Login failed: ' + error.message, 'error');
  } finally {
    loginBtn.disabled = false;
    loginBtn.textContent = 'Login';
  }
}

// Handle logout
function handleLogout() {
  if (!confirm('Are you sure you want to logout?')) {
    return;
  }

  chrome.runtime.sendMessage({ action: 'LOGOUT' }, () => {
    loadUserProfile();
    showMessage('Logged out successfully', 'success');
  });
}

// Save all settings
function saveSettings() {
  const settings = {
    notificationsEnabled: document.getElementById('notifications-enabled').checked,
    notifyAchievements: document.getElementById('notify-achievements').checked,
    notifyReminders: document.getElementById('notify-reminders').checked,
    notifyMilestones: document.getElementById('notify-milestones').checked,
    syncInterval: parseInt(document.getElementById('sync-interval').value),
    autoSync: document.getElementById('auto-sync').checked,
    floatingToolbar: document.getElementById('floating-toolbar').checked,
    theme: document.getElementById('theme').value,
    apiUrl: document.getElementById('api-url').value || DEFAULT_SETTINGS.apiUrl,
    debugMode: document.getElementById('debug-mode').checked
  };

  chrome.storage.sync.set(settings, () => {
    // Notify background script of changes
    chrome.runtime.sendMessage({
      action: 'SETTINGS_CHANGED',
      settings
    });

    showMessage('Settings saved successfully', 'success');
  });
}

// Clear browser cache
function clearCache() {
  if (!confirm('Clear all cached data? This cannot be undone.')) {
    return;
  }

  chrome.storage.local.clear(() => {
    showMessage('Cache cleared', 'success');
  });
}

// Reset to default settings
function resetToDefaults() {
  if (!confirm('Reset all settings to defaults? This cannot be undone.')) {
    return;
  }

  chrome.storage.sync.set(DEFAULT_SETTINGS, () => {
    loadSettings();
    showMessage('Settings reset to defaults', 'success');
  });
}

// Show status message
function showMessage(message, type = 'info') {
  const statusEl = document.getElementById('status-message');
  statusEl.textContent = message;
  statusEl.className = `status-message ${type}`;

  setTimeout(() => {
    statusEl.className = 'status-message';
  }, 3000);
}

console.log('[EduApps] Options script loaded');
