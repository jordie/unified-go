/**
 * EduApps Chrome Extension - Popup Script
 * Handles popup UI interactions and state management
 */

const API_BASE_URL = 'https://localhost:5051';

// DOM Elements
const loginForm = document.getElementById('login-form');
const userInfo = document.getElementById('user-info');
const usernameInput = document.getElementById('username');
const passwordInput = document.getElementById('password');
const loginBtn = document.getElementById('login-btn');
const logoutBtn = document.getElementById('logout-btn');
const userName = document.getElementById('user-name');
const statsSection = document.getElementById('stats-section');
const practiceSection = document.getElementById('practice-section');
const dashboardSection = document.getElementById('dashboard-section');
const dashboardLink = document.getElementById('dashboard-link');
const notificationsToggle = document.getElementById('notifications-toggle');
const syncIntervalSelect = document.getElementById('sync-interval');

// Initialize on popup load
document.addEventListener('DOMContentLoaded', () => {
  loadUserState();
  loadStats();
  attachEventListeners();
  loadSettings();
});

// Load user state from storage
function loadUserState() {
  chrome.storage.sync.get(['userId', 'username'], (data) => {
    if (data.userId && data.username) {
      showUserInfo(data.username);
      setupDashboardLink(data.userId);
    } else {
      showLoginForm();
    }
  });
}

// Show login form
function showLoginForm() {
  loginForm.style.display = 'flex';
  userInfo.style.display = 'none';
  statsSection.style.display = 'none';
  practiceSection.style.display = 'none';
  dashboardSection.style.display = 'none';
}

// Show user info
function showUserInfo(username) {
  loginForm.style.display = 'none';
  userInfo.style.display = 'flex';
  statsSection.style.display = 'block';
  practiceSection.style.display = 'block';
  dashboardSection.style.display = 'block';

  userName.textContent = username;

  // Set avatar emoji based on first letter
  const avatar = document.getElementById('user-avatar');
  const firstLetter = username.charAt(0).toUpperCase();
  avatar.textContent = firstLetter;
}

// Setup dashboard link
function setupDashboardLink(userId) {
  dashboardLink.href = `${API_BASE_URL}/dashboard?user_id=${userId}`;
}

// Load stats from storage
function loadStats() {
  chrome.storage.sync.get(['stats'], (data) => {
    const stats = data.stats || {};

    document.getElementById('typing-sessions').textContent = stats.typingSessions || 0;
    document.getElementById('math-sessions').textContent = stats.mathSessions || 0;
    document.getElementById('reading-sessions').textContent = stats.readingSessions || 0;
    document.getElementById('total-minutes').textContent = stats.totalMinutes || 0;
  });
}

// Load settings from storage
function loadSettings() {
  chrome.storage.sync.get(['notificationsEnabled', 'syncInterval'], (data) => {
    notificationsToggle.checked = data.notificationsEnabled !== false;
    syncIntervalSelect.value = data.syncInterval || 5;
  });
}

// Attach event listeners
function attachEventListeners() {
  // Login
  loginBtn.addEventListener('click', handleLogin);
  usernameInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') handleLogin();
  });
  passwordInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') handleLogin();
  });

  // Logout
  logoutBtn.addEventListener('click', handleLogout);

  // Quick practice buttons
  document.querySelectorAll('.quick-btn').forEach(btn => {
    btn.addEventListener('click', (e) => {
      e.preventDefault();
      const practiceType = btn.dataset.practice;
      chrome.tabs.create({
        url: `${API_BASE_URL}/${practiceType}`
      });
    });
  });

  // Settings
  notificationsToggle.addEventListener('change', (e) => {
    chrome.storage.sync.set({
      notificationsEnabled: e.target.checked
    });
  });

  syncIntervalSelect.addEventListener('change', (e) => {
    chrome.storage.sync.set({
      syncInterval: parseInt(e.target.value)
    });
  });

  // Footer links
  document.getElementById('help-link').addEventListener('click', (e) => {
    e.preventDefault();
    chrome.tabs.create({ url: `${API_BASE_URL}/help` });
  });

  document.getElementById('feedback-link').addEventListener('click', (e) => {
    e.preventDefault();
    chrome.tabs.create({ url: `${API_BASE_URL}/feedback` });
  });

  document.getElementById('settings-link').addEventListener('click', (e) => {
    e.preventDefault();
    chrome.runtime.openOptionsPage();
  });

  // Refresh stats periodically
  setInterval(loadStats, 30000); // Every 30 seconds
}

// Handle login
async function handleLogin() {
  const username = usernameInput.value.trim();
  const password = passwordInput.value.trim();

  if (!username || !password) {
    showError('Please enter username and password');
    return;
  }

  loginBtn.disabled = true;
  loginBtn.textContent = 'Logging in...';

  try {
    // Call login API
    const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
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
      // Clear form
      usernameInput.value = '';
      passwordInput.value = '';

      // Update UI
      showUserInfo(data.username);
      setupDashboardLink(data.userId);

      // Trigger initial sync
      chrome.runtime.sendMessage({
        action: 'SYNC_DATA'
      });

      showSuccess('Login successful!');
    });
  } catch (error) {
    console.error('Login error:', error);
    showError('Login failed: ' + error.message);
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
    // Clear form
    usernameInput.value = '';
    passwordInput.value = '';

    // Update UI
    showLoginForm();
    showSuccess('Logged out successfully');
  });
}

// Show error message
function showError(message) {
  const errorDiv = document.createElement('div');
  errorDiv.className = 'error';
  errorDiv.textContent = message;

  const authSection = document.getElementById('auth-section');
  authSection.insertBefore(errorDiv, authSection.firstChild);

  setTimeout(() => {
    errorDiv.remove();
  }, 3000);
}

// Show success message
function showSuccess(message) {
  const successDiv = document.createElement('div');
  successDiv.className = 'success';
  successDiv.textContent = message;

  const authSection = document.getElementById('auth-section');
  authSection.insertBefore(successDiv, authSection.firstChild);

  setTimeout(() => {
    successDiv.remove();
  }, 3000);
}

console.log('[EduApps] Popup script loaded');
