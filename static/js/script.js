"use strict";

// TODO:
// - hijack ctrl-s
const saveButton = document.getElementById("header__save");
saveButton.addEventListener(
	"click",
	async function () {
		const textAreaValue = document.getElementById("main__input");
		console.log(textAreaValue.value);

		const response = await fetch("/snippets", {
			method: "POST",
			body: JSON.stringify({
				snippet: { text: textAreaValue.value },
			}),
		});

		const data = await response.json();
		const addr = data.snippet.addr;

		// NOTE: we can use window.history.pushState to avoid refreshing page
		// but it won't update the go template
		window.location.replace(addr);
	},
	{ once: true }
);
