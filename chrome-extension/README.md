# EduApps Chrome Extension

A powerful learning companion extension for the EduApps educational platform. Practice typing, math, and reading skills seamlessly across the web.

## Features

### Core Features
- **Quick Practice Access**: One-click access to typing, math, and reading practice
- **Real-time Sync**: Automatically sync your progress with the main EduApps platform
- **Statistics Tracking**: View today's practice statistics in the extension popup
- **User Authentication**: Secure login directly from the extension
- **Settings Management**: Configure notifications, sync intervals, and more

### Content Integration
- **Floating Toolbar**: Quick access buttons injected into any webpage
- **Page Analysis**: Automatically detect reading and math content on pages
- **Text Selection**: Right-click selected text for instant practice
- **Keyboard Shortcuts**:
  - `Ctrl+Shift+E` (Cmd+Shift+E on Mac): Toggle extension
  - `Ctrl+Shift+T` (Cmd+Shift+T on Mac): Quick typing practice
  - `Ctrl+Shift+M` (Cmd+Shift+M on Mac): Quick math practice

### Smart Features
- **Omnibox Integration**: Type "eduapps [keyword]" in the address bar
- **Web API**: Pages can access `window.EduApps` for content analysis
- **Background Sync**: Periodic background synchronization of stats
- **Notifications**: Achievement and reminder notifications

## Installation

### For Development

1. Clone or extract the extension files to a folder
2. Open Chrome and navigate to `chrome://extensions/`
3. Enable "Developer mode" (toggle in top right)
4. Click "Load unpacked"
5. Select the `chrome-extension` directory

### For Production

The extension will be available on the Chrome Web Store (coming soon).

## File Structure

```
chrome-extension/
├── manifest.json           # Extension configuration
├── background.js          # Service worker for background tasks
├── content.js            # Content script for page injection
├── injected.js           # Page context injection script
├── popup.html            # Extension popup UI
├── popup.js              # Popup interaction logic
├── popup.css             # Popup styling
├── options.html          # Options page (optional)
├── onboarding.html       # First-run onboarding (optional)
├── icons/                # Extension icons
│   ├── icon-16.png
│   ├── icon-48.png
│   └── icon-128.png
└── README.md            # This file
```

## Configuration

### Manifest Settings

**Key Permissions:**
- `activeTab`: Access to current tab
- `scripting`: Execute scripts on pages
- `storage`: Store user data locally
- `notifications`: Send notifications
- `webRequest`: Monitor network requests
- `tabs`: Tab management
- `webNavigation`: Monitor page navigation

**Host Permissions:**
- `<all_urls>`: Access to all websites for content analysis

## Usage

### For Users

1. **Login**: Click the extension icon and enter your EduApps credentials
2. **Quick Practice**: Use the buttons in the popup to access typing, math, or reading practice
3. **View Stats**: Check today's progress in the main popup view
4. **Settings**: Configure notifications and sync preferences
5. **Floating Toolbar**: Enable/disable the floating toolbar in settings

### For Developers

The extension exposes a global API for pages to use:

```javascript
// Get selected text
const text = window.EduApps.getSelectedText();

// Get page content analysis
const analysis = window.EduApps.getPageContent();

// Check content type
if (window.EduApps.isReadingContent()) {
  console.log('This page has reading content');
}

if (window.EduApps.isMathContent()) {
  console.log('This page has math content');
}

// Extract form data
const forms = window.EduApps.getFormData();

// Get page metadata
const metadata = window.EduApps.getMetadata();

// Log activity (sends to background script)
window.EduApps.logActivity({
  type: 'page_view',
  content_type: 'reading',
  duration: 120
});

// Track user interactions
window.EduApps.trackInteractions((event) => {
  console.log('User interaction:', event);
});
```

## Background Script API

The background script handles all server communication. Content scripts can send messages:

```javascript
// Get current user
chrome.runtime.sendMessage({ action: 'GET_USER' }, (response) => {
  console.log('User:', response.userId, response.username);
});

// Sync data with server
chrome.runtime.sendMessage({ action: 'SYNC_DATA' }, (response) => {
  console.log('Synced:', response.data);
});

// Log a practice session
chrome.runtime.sendMessage({
  action: 'LOG_SESSION',
  app: 'typing',
  duration: 300,
  score: 85,
  correctCount: 42,
  totalCount: 50
}, (response) => {
  console.log('Session logged');
});

// Get statistics
chrome.runtime.sendMessage({ action: 'GET_STATS' }, (response) => {
  console.log('Stats:', response.stats);
});

// Show notification
chrome.runtime.sendMessage({
  action: 'SHOW_NOTIFICATION',
  title: 'Great job!',
  message: 'You earned a badge!',
  iconUrl: chrome.runtime.getURL('icons/icon-128.png'),
  clickUrl: 'https://localhost:5051/dashboard'
});

// User login
chrome.runtime.sendMessage({
  action: 'SET_USER',
  userId: 123,
  apiToken: 'token...',
  username: 'john_doe'
}, (response) => {
  console.log('User set:', response.success);
});

// User logout
chrome.runtime.sendMessage({ action: 'LOGOUT' }, (response) => {
  console.log('Logged out');
});
```

## Storage

The extension uses Chrome's `storage.sync` API for persistent data:

```javascript
// Stored data structure
{
  extensionEnabled: true,
  userId: 123,
  username: 'john_doe',
  apiToken: 'token...',
  syncInterval: 5,
  notificationsEnabled: true,
  lastSync: '2026-02-21T14:00:00Z',
  lastLogin: '2026-02-21T10:00:00Z',
  stats: {
    typingSessions: 5,
    mathSessions: 3,
    readingSessions: 2,
    totalMinutes: 45
  }
}
```

## Configuration Options

**Sync Interval**: How often to sync stats with the server
- Options: 5, 10, 30, 60 minutes
- Default: 5 minutes

**Notifications**: Enable/disable achievement and reminder notifications
- Default: Enabled

**API Endpoint**: Configure the EduApps server URL
- Default: `https://localhost:5051`
- Can be changed in `background.js` and `content.js`

## Security

### HTTPS Only
- Extension communicates with EduApps over HTTPS
- All API tokens are stored securely in Chrome's storage API
- Tokens are never exposed to web content

### Content Security Policy
- Scripts are properly isolated
- No inline scripts or dangerous operations
- Input validation on all forms

### Permissions
- Only necessary permissions are requested
- Users can revoke access at any time via Chrome settings

## Keyboard Shortcuts

| Shortcut | Action | Platform |
|----------|--------|----------|
| `Ctrl+Shift+E` | Toggle extension | Windows/Linux |
| `Cmd+Shift+E` | Toggle extension | macOS |
| `Ctrl+Shift+T` | Quick typing | Windows/Linux |
| `Cmd+Shift+T` | Quick typing | macOS |
| `Ctrl+Shift+M` | Quick math | Windows/Linux |
| `Cmd+Shift+M` | Quick math | macOS |

## Troubleshooting

### Extension not loading
1. Check that all files are in the extension directory
2. Verify manifest.json is valid JSON
3. Check browser console for error messages
4. Clear browser cache and reload

### Authentication failing
1. Verify EduApps server is running
2. Check that API endpoint URL is correct
3. Verify credentials are correct
4. Check browser console for network errors

### Sync not working
1. Verify internet connection
2. Check that EduApps server is accessible
3. Verify API token is valid
4. Check sync interval setting

### Icons not showing
1. Verify icon files exist in `icons/` directory
2. Check icon file names match manifest
3. Verify icon formats (PNG, square aspect ratio)
4. Reload extension in `chrome://extensions/`

## Development

### Building from Source

1. Make changes to files
2. Go to `chrome://extensions/`
3. Click the reload button for the extension
4. Changes take effect immediately

### Testing

1. Test login/logout
2. Test quick practice buttons
3. Test statistics sync
4. Test notifications
5. Test keyboard shortcuts
6. Test on different websites
7. Test offline behavior
8. Check console for errors

## Future Enhancements

- [ ] Voice commands for hands-free practice
- [ ] Screenshot recognition for math problems
- [ ] Peer collaboration features
- [ ] Achievement system with badges
- [ ] Custom practice plans
- [ ] Integration with learning management systems
- [ ] Offline mode with sync on reconnect
- [ ] Multiple language support

## Support

For issues, feature requests, or feedback:
- Email: support@eduapps.com
- GitHub: https://github.com/jgirmay/unified-go/issues
- Web: https://eduapps.com/support

## License

This extension is part of the EduApps educational platform. All rights reserved.

## Version History

### v1.0.0 (2026-02-21)
- Initial release
- Core features: typing, math, reading practice
- Real-time sync with EduApps server
- Floating toolbar injection
- Statistics tracking

---

**Made with ❤️ by the EduApps Team**
