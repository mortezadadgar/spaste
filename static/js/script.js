"use strict";

const saveButton = document.querySelector(".js-header__save");
saveButton.addEventListener(
	"click",
	async function () {
		const inpuText = document.querySelector(".js-main__input");
		const data = inpuText.value;

		const resp = await fetch("/snippets", {
			method: "POST",
			body: JSON.stringify({ snippet : { data: data }}),
		});

		if (!resp.ok) {
			console.log("POST /snippets request was not ok");
		}
	},
	{ once: true }
);
