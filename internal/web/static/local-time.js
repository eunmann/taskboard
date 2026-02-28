// Converts <time data-local> elements from UTC to local timezone.
function convertLocalTimes() {
    document.querySelectorAll('time[data-local]').forEach(function(el) {
        var dt = el.getAttribute('datetime');
        if (!dt) return;

        var date = new Date(dt);
        if (isNaN(date.getTime())) return;

        var options = {
            year: 'numeric', month: 'short', day: 'numeric',
            hour: 'numeric', minute: '2-digit',
            timeZoneName: 'short'
        };

        el.textContent = date.toLocaleDateString('en-US', options);
    });
}

document.addEventListener('DOMContentLoaded', convertLocalTimes);
document.body.addEventListener('htmx:afterSwap', convertLocalTimes);
