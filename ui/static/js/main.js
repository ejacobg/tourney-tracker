// Render all HTTP responses, even errors.
// https://htmx.org/events/#htmx:beforeSwap
// https://htmx.org/docs/#modifying_swapping_behavior_with_events
document.body.addEventListener('htmx:beforeSwap', function (event) {
    event.detail.shouldSwap = true;
});