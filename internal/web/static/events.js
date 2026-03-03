// events.js — SSE client for live reactor events
(function() {
    'use strict';

    var es = null;
    var toastStack = [];
    var MAX_TOASTS = 3;

    function connect() {
        if (es) es.close();

        es = new EventSource('/stream');

        es.addEventListener('reactor', function(e) {
            try {
                var event = JSON.parse(e.data);
                showToast(event);
                triggerRefresh(event);
            } catch (err) {
                // ignore parse errors
            }
        });

        es.addEventListener('open', function() {
            var indicator = document.getElementById('sse-status');
            if (indicator) {
                indicator.className = 'sse-indicator sse-connected';
                indicator.title = 'Live updates connected';
            }
        });

        es.addEventListener('error', function() {
            var indicator = document.getElementById('sse-status');
            if (indicator) {
                indicator.className = 'sse-indicator sse-disconnected';
                indicator.title = 'Live updates reconnecting...';
            }
        });
    }

    function showToast(event) {
        var text = event.summary || event.event_type;
        if (!text) return;

        // Limit toast stack
        while (toastStack.length >= MAX_TOASTS) {
            var old = toastStack.shift();
            if (old.parentNode) old.parentNode.removeChild(old);
        }

        var toast = document.createElement('div');
        toast.className = 'sse-toast';

        var badge = document.createElement('span');
        badge.className = 'sse-toast-type';
        badge.textContent = event.event_type || 'event';
        toast.appendChild(badge);

        var msg = document.createElement('span');
        msg.className = 'sse-toast-msg';
        msg.textContent = text;
        toast.appendChild(msg);

        var container = document.getElementById('sse-toasts');
        if (!container) {
            container = document.createElement('div');
            container.id = 'sse-toasts';
            document.body.appendChild(container);
        }
        container.appendChild(toast);
        toastStack.push(toast);

        // Animate in
        setTimeout(function() { toast.classList.add('sse-toast-show'); }, 10);

        // Fade out after 5s
        setTimeout(function() {
            toast.classList.remove('sse-toast-show');
            setTimeout(function() {
                if (toast.parentNode) toast.parentNode.removeChild(toast);
                var idx = toastStack.indexOf(toast);
                if (idx !== -1) toastStack.splice(idx, 1);
            }, 300);
        }, 5000);
    }

    function triggerRefresh(event) {
        // Fire sse-refresh on body — HTMX pages listen via "from:body"
        if (typeof htmx !== 'undefined') {
            htmx.trigger(document.body, 'sse-refresh');
        }
    }

    // Connect when DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', connect);
    } else {
        connect();
    }
})();
