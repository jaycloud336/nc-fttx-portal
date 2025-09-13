// Minimal JavaScript for NC FTTX Portal
// Added by developers for basic interactivity

document.addEventListener('DOMContentLoaded', function() {
    console.log('NC FTTX Portal loaded successfully');
    
    // Add simple analytics tracking for DevOps metrics
    trackPageView();
    
    // Add click tracking for external links
    addLinkTracking();
});

function trackPageView() {
    // Simple page view tracking (would integrate with monitoring)
    const pageData = {
        page: window.location.pathname,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent
    };
    
    // In production, this would send to analytics/monitoring system
    console.log('Page view:', pageData);
}

function addLinkTracking() {
    // Track clicks on external links (useful for monitoring)
    const externalLinks = document.querySelectorAll('a[target="_blank"]');
    
    externalLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            const linkData = {
                url: this.href,
                text: this.textContent.trim(),
                timestamp: new Date().toISOString()
            };
            
            // In production, this would send to monitoring system
            console.log('External link clicked:', linkData);
        });
    });
}

// Health check function for monitoring
function healthCheck() {
    return fetch('/health')
        .then(response => response.json())
        .then(data => {
            console.log('Health check:', data);
            return data;
        })
        .catch(error => {
            console.error('Health check failed:', error);
            return { status: 'error', error: error.message };
        });
}

// Make health check available globally for testing
window.ncFttxPortal = {
    healthCheck: healthCheck
};
