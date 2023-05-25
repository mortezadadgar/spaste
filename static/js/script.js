"use strict";

const saveButton = document.getElementById("header__save");
saveButton.addEventListener(
	"click",
	async function () {
		const textAreaValue = document.getElementById("main__input");
		const headerInput = document.getElementById("header__input");

		const response = await fetch("/snippets", {
			method: "POST",
			body: JSON.stringify({
				snippet: { text: textAreaValue.getAttribute("value") },
			}),
		});

		const data = await response.json();
		const addr = data.snippet.addr;

		window.location.replace(addr);

		headerInput.setAttribute("value", `${window.location.host}/${addr}`);
	},
	// TODO: new pages get new event
	{ once: true }
);
