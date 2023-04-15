// If an error response, use the hx-swap behavior of the #error element.
// hx-swap uses the value of the AJAX target, not the #error.
// https://htmx.org/attributes/hx-swap/
// document.body.addEventListener('htmx:afterOnLoad', function (event) {
//     if (event.detail.failed) {
//         event.detail.target = htmx.find('#error')
//     }
// })

// Render all HTTP responses, even errors.
// https://htmx.org/events/#htmx:beforeSwap
// https://htmx.org/docs/#modifying_swapping_behavior_with_events
document.body.addEventListener('htmx:beforeSwap', function (event) {
    event.detail.shouldSwap = true;

    // If the response is a non-200 code, then render it in the #error element.
    // Note: retargeting doesn't seem to respect the new target's hx-swap attribute.
    // Instead, it will use the hx-swap of the original target, rather than the new one.
    // if (event.detail.isError) {
    //     event.detail.target = htmx.find('#error')
    // }

    // SOLVED: using the HX-Reswap and HX-Retarget headers.
});

// You may also use the htmx helper: htmx.on('<event>', function(event) {...})
// https://htmx.org/docs/#events

