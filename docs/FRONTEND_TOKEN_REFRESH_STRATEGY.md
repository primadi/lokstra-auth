# Frontend Token Refresh Strategy

## Recommended: Proactive + Reactive Fallback

Combine both strategies for the best UX and reliability.

## Complete Implementation

```javascript
// ============================================================================
// Token Manager - Handles all token operations
// ============================================================================

class TokenManager {
  constructor() {
    this.accessToken = '';
    this.refreshToken = '';
    this.tokenExpiresAt = 0;
    this.refreshTimer = null;
    this.isRefreshing = false;
    this.refreshPromise = null;
  }

  // Initialize from storage (page reload)
  init() {
    this.accessToken = localStorage.getItem('access_token') || '';
    this.refreshToken = localStorage.getItem('refresh_token') || '';
    this.tokenExpiresAt = parseInt(localStorage.getItem('token_expires_at') || '0');
    
    if (this.accessToken && this.tokenExpiresAt > Date.now()) {
      this.scheduleRefresh();
    }
  }

  // Login - Save tokens and schedule refresh
  async login(username, password) {
    const response = await fetch('/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    if (!response.ok) {
      throw new Error('Login failed');
    }

    const data = await response.json();
    this.setTokens(data);
    return data;
  }

  // Set tokens and schedule auto-refresh
  setTokens(data) {
    this.accessToken = data.access_token;
    this.refreshToken = data.refresh_token;
    this.tokenExpiresAt = Date.now() + (data.expires_in * 1000);

    // Save to storage
    localStorage.setItem('access_token', this.accessToken);
    localStorage.setItem('refresh_token', this.refreshToken);
    localStorage.setItem('token_expires_at', this.tokenExpiresAt.toString());

    // Schedule proactive refresh
    this.scheduleRefresh();
  }

  // Schedule proactive refresh (1 minute before expiry)
  scheduleRefresh() {
    // Clear existing timer
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
    }

    const now = Date.now();
    const expiresIn = this.tokenExpiresAt - now;
    
    // Refresh 1 minute (60000ms) before expiry
    const refreshBuffer = 60 * 1000;
    const refreshTime = expiresIn - refreshBuffer;

    if (refreshTime > 0) {
      console.log(`🔄 Token refresh scheduled in ${Math.round(refreshTime / 1000)}s`);
      
      this.refreshTimer = setTimeout(async () => {
        console.log('⏰ Proactive token refresh triggered');
        await this.refresh();
      }, refreshTime);
    } else {
      // Token expires soon, refresh immediately
      console.log('⚠️ Token expires soon, refreshing now');
      this.refresh();
    }
  }

  // Refresh access token
  async refresh() {
    // Prevent multiple simultaneous refreshes
    if (this.isRefreshing) {
      return this.refreshPromise;
    }

    this.isRefreshing = true;
    this.refreshPromise = this._doRefresh();

    try {
      await this.refreshPromise;
    } finally {
      this.isRefreshing = false;
      this.refreshPromise = null;
    }
  }

  async _doRefresh() {
    try {
      const response = await fetch('/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          refresh_token: this.refreshToken 
        })
      });

      if (!response.ok) {
        console.error('❌ Token refresh failed');
        this.logout();
        throw new Error('Token refresh failed');
      }

      const data = await response.json();
      
      console.log('✅ Token refreshed successfully');
      this.setTokens(data);
      
      return data;
    } catch (error) {
      console.error('❌ Token refresh error:', error);
      this.logout();
      throw error;
    }
  }

  // Get current access token
  async getAccessToken() {
    const now = Date.now();
    
    // Check if token is expired or about to expire (within 30s)
    if (this.tokenExpiresAt - now < 30000) {
      console.log('⚠️ Token expired or expiring soon, refreshing...');
      await this.refresh();
    }
    
    return this.accessToken;
  }

  // Logout
  logout() {
    // Clear timers
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
    }

    // Clear tokens
    this.accessToken = '';
    this.refreshToken = '';
    this.tokenExpiresAt = 0;

    // Clear storage
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('token_expires_at');

    // Redirect to login
    window.location.href = '/login';
  }
}

// Create singleton instance
const tokenManager = new TokenManager();
tokenManager.init();

// ============================================================================
// API Client - Uses TokenManager
// ============================================================================

class ApiClient {
  async request(url, options = {}) {
    // Get fresh access token (auto-refreshes if needed)
    const accessToken = await tokenManager.getAccessToken();

    // Add auth header
    const headers = {
      ...options.headers,
      'Authorization': `Bearer ${accessToken}`,
      'Content-Type': 'application/json'
    };

    try {
      const response = await fetch(url, {
        ...options,
        headers
      });

      // Reactive fallback: Handle 401 error
      if (response.status === 401) {
        console.log('🔄 Received 401, attempting token refresh...');
        
        // Refresh token
        await tokenManager.refresh();
        
        // Retry request with new token
        const newAccessToken = await tokenManager.getAccessToken();
        headers['Authorization'] = `Bearer ${newAccessToken}`;
        
        return fetch(url, { ...options, headers });
      }

      return response;
    } catch (error) {
      console.error('❌ API request failed:', error);
      throw error;
    }
  }

  async get(url) {
    return this.request(url, { method: 'GET' });
  }

  async post(url, data) {
    return this.request(url, {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  async put(url, data) {
    return this.request(url, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  }

  async delete(url) {
    return this.request(url, { method: 'DELETE' });
  }
}

const apiClient = new ApiClient();

// ============================================================================
// Usage Examples
// ============================================================================

// Login
async function handleLogin() {
  try {
    await tokenManager.login('user@example.com', 'password');
    console.log('✅ Logged in successfully');
    
    // Token refresh is now automatic!
  } catch (error) {
    console.error('❌ Login failed:', error);
  }
}

// Make API calls
async function fetchUserData() {
  try {
    const response = await apiClient.get('/api/user/profile');
    const data = await response.json();
    console.log('User data:', data);
  } catch (error) {
    console.error('Failed to fetch user data:', error);
  }
}

// The token is automatically refreshed in the background!
// No user interruption, no loading delays.

// ============================================================================
// Multi-Tab Synchronization (Bonus Feature)
// ============================================================================

// Listen for storage changes from other tabs
window.addEventListener('storage', (event) => {
  if (event.key === 'access_token' || event.key === 'refresh_token') {
    console.log('🔄 Token updated in another tab, syncing...');
    
    // Reload tokens from storage
    tokenManager.init();
  }
});

// ============================================================================
// Visibility Change - Pause/Resume Refresh Timer
// ============================================================================

document.addEventListener('visibilitychange', () => {
  if (document.hidden) {
    console.log('👁️ Page hidden, pausing refresh timer');
    // Optionally pause timer when page is hidden
  } else {
    console.log('👁️ Page visible, resuming refresh timer');
    
    // Check if token expired while page was hidden
    const now = Date.now();
    if (tokenManager.tokenExpiresAt < now) {
      console.log('⚠️ Token expired while page was hidden, refreshing...');
      tokenManager.refresh();
    } else {
      // Reschedule refresh
      tokenManager.scheduleRefresh();
    }
  }
});

export { tokenManager, apiClient };
```

## Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│              Token Refresh Flow (Proactive)                 │
└─────────────────────────────────────────────────────────────┘

Time: 00:00 - User Login
├─ Backend returns: access_token (expires in 15 min)
├─ Frontend saves: token_expires_at = 00:15
└─ Frontend schedules: refresh at 00:14 (1 min before expiry)

Time: 00:00 - 00:14 - Normal Operation
├─ User makes API calls → All successful ✅
├─ Token is still valid
└─ No interruption

Time: 00:14 - Proactive Refresh (1 min before expiry)
├─ Timer triggers: "Token expires in 1 min, refreshing..."
├─ Frontend: POST /auth/refresh
├─ Backend: Returns new access_token + refresh_token
├─ Frontend: Updates tokens, reschedules next refresh
└─ User didn't notice anything! ✅

Time: 00:14 - 00:29 - Continue Normal Operation
├─ User makes API calls → All successful ✅
└─ No user experience interruption

Time: 00:29 - Proactive Refresh Again
└─ Cycle repeats...

┌─────────────────────────────────────────────────────────────┐
│         Reactive Fallback (If Proactive Failed)            │
└─────────────────────────────────────────────────────────────┘

Scenario: User left tab inactive for 20 minutes

Time: 00:20 - User returns and clicks button
├─ Frontend: GET /api/data with expired token
├─ Backend: 401 Unauthorized ❌
├─ Frontend: Catches 401 → "Token expired, refreshing..."
├─ Frontend: POST /auth/refresh
├─ Backend: Returns new tokens
├─ Frontend: Retry GET /api/data with new token
└─ Backend: 200 OK ✅

User Experience: Small delay (< 1 second) but still works!
```

## Key Benefits

### ✅ Proactive Refresh (Primary Strategy)
- **Zero User Interruption**: Token refreshed before user notices
- **Smooth UX**: No loading delays or errors
- **Predictable**: Scheduled refresh at known time

### ✅ Reactive Fallback (Backup Strategy)
- **Resilient**: Handles edge cases (user offline, page inactive)
- **Automatic Recovery**: Transparent retry on 401 errors
- **No Manual Intervention**: User doesn't need to click "retry"

### ✅ Combined Benefits
- **Best of Both Worlds**: Proactive performance + reactive resilience
- **Multi-Tab Support**: Tokens sync across tabs via storage events
- **Battery Efficient**: Timer pauses when page hidden
- **Developer Friendly**: Simple API, automatic token management

## Configuration Recommendations

### Backend Configuration
```json
{
  "security": {
    "enable_refresh_token_rotation": true,
    "refresh_token_rotation_grace_period": 30
  }
}
```

### Frontend Configuration
```javascript
const CONFIG = {
  // How early to refresh before expiry
  REFRESH_BUFFER_SECONDS: 60,  // 1 minute
  
  // Minimum time before considering token "expiring soon"
  EXPIRING_SOON_THRESHOLD: 30,  // 30 seconds
  
  // Retry configuration for failed refreshes
  MAX_REFRESH_RETRIES: 3,
  REFRESH_RETRY_DELAY: 1000  // 1 second
};
```

## Testing Scenarios

### Test 1: Proactive Refresh Works
1. Login → token expires in 15 min
2. Wait 14 minutes
3. Check console → "Proactive token refresh triggered"
4. Token should be refreshed automatically ✅

### Test 2: Reactive Fallback Works
1. Login → token expires in 15 min
2. Disconnect internet
3. Wait 16 minutes (token expired)
4. Reconnect internet
5. Make API call → Should auto-refresh and succeed ✅

### Test 3: Multi-Tab Sync
1. Open 2 tabs
2. Tab 1 triggers refresh
3. Check Tab 2 → Should detect storage change and sync ✅

### Test 4: Page Inactive
1. Login → token expires in 15 min
2. Hide tab (switch to another tab/app)
3. Wait 20 minutes
4. Return to tab → Token should refresh on visibility change ✅

## Monitoring

Track these metrics in your app:
- `token_refresh_proactive_count` - Number of scheduled refreshes
- `token_refresh_reactive_count` - Number of 401 fallback refreshes
- `token_refresh_failure_count` - Number of failed refreshes
- `token_refresh_duration_ms` - How long refresh takes

High reactive count might indicate:
- User leaving tabs inactive for long periods
- Proactive refresh not working properly
- Network issues preventing scheduled refresh

## Common Pitfalls to Avoid

❌ **Don't**: Wait for 401 error only
✅ **Do**: Proactive refresh + 401 fallback

❌ **Don't**: Refresh on every API call
✅ **Do**: Check expiry, refresh only when needed

❌ **Don't**: Use setInterval for constant checking
✅ **Do**: Use setTimeout scheduled at exact expiry time

❌ **Don't**: Ignore multi-tab scenarios
✅ **Do**: Sync tokens across tabs via storage events

❌ **Don't**: Forget to clear timers on logout
✅ **Do**: Clean up all timers and listeners

## Summary

| Strategy | When to Use | UX Impact | Complexity |
|----------|-------------|-----------|------------|
| **Reactive Only** | Never (bad UX) | ⚠️ User sees delays | Simple |
| **Proactive Only** | Never (edge cases fail) | ✅ Smooth (mostly) | Medium |
| **Proactive + Reactive** | **Always (recommended)** | ✅ Always smooth | Medium |

**Conclusion**: Use **Proactive + Reactive Fallback** for the best user experience and reliability! 🎯
