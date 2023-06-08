"use strict";

// TODO:
// - hijack ctrl-s

// throw "wait";

const mainInput = document.querySelector(".main__input");
const saveButton = document.querySelector(".header__save");
const headerSelect = document.querySelector(".header__select");
let lineCount = 0;

saveButton.addEventListener(
	"click",
	async function () {
		const response = await fetch("/snippets", {
			method: "POST",
			body: JSON.stringify({
				snippet: {
					text: mainInput.value,
					lang: headerSelect.value,
					lineCount: lineCount,
				},
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

const mainNumbers = document.querySelector(".main__numbers");
const mainNumber = document.querySelector(".main__number");

mainInput.addEventListener("input", function () {
	mainNumbers.innerHTML = "";
	var textLines = mainInput.value.split("\n");
	for (let i = 0; i < textLines.length; i++) {
		const numberElement = mainNumber.cloneNode(true);
		numberElement.innerText = i + 1;
		mainNumbers.appendChild(numberElement);
	}
	lineCount = textLines.length;
});

mainInput.addEventListener("scroll", function () {
	mainNumbers.scrollTo(0, mainInput.scrollTop);
});
