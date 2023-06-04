"use strict";

// TODO:
// - hijack ctrl-s
const saveButton = document.querySelector(".header__save");
saveButton.addEventListener(
	"click",
	async function () {
		const mainInput = document.querySelector(".main__input");
		const headerSelect = document.querySelector(".header__select");

		const response = await fetch("/snippets", {
			method: "POST",
			body: JSON.stringify({
				snippet: { text: mainInput.value, lang: headerSelect.value },
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
