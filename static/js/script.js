"use strict";

const storage = window.localStorage;

if (storage.getItem("light-theme") == "true") {
	document.querySelector("html").classList.toggle("light-theme");
}

const themeButton = document.querySelector(".header__theme");

themeButton.addEventListener("click", function () {
	if (storage.getItem("light-theme") == "true") {
		themeButton.src = "./static/svg/sun.svg";
	} else {
		themeButton.src = "./static/svg/moon.svg";
	}
});

themeButton.addEventListener("click", function () {
	if (storage.getItem("light-theme") == "true") {
		storage.setItem("light-theme", false);
	} else {
		storage.setItem("light-theme", true);
	}

	document.querySelector("html").classList.toggle("light-theme");
});

const saveButton = document.querySelector(".header__save");

window.addEventListener("keydown", function (event) {
	if (event.key == "s" && event.ctrlKey) {
		event.preventDefault();
		saveButton.click();
	}
});

const shareButton = document.querySelector(".header__share");
const snippetAddr = document.querySelector(".header__input");

shareButton.addEventListener("click", function () {
	if (snippetAddr.value.length == 0) {
		alert("Please save your code first.");
		return;
	}

	navigator.clipboard.writeText(snippetAddr.value);
});

const snippetText = document.querySelector(".main__input");
const SnippetSelect = document.querySelector(".header__select");
let lineCount = 0;

if (saveButton != undefined) {
	saveButton.addEventListener("click", async function () {
		if (snippetText.value.length < 1) {
			alert("Please paste your code.");
			return;
		}

		const response = await fetch("/snippets", {
			method: "POST",
			body: JSON.stringify({
				snippet: {
					text: snippetText.value,
					lang: SnippetSelect.value,
					lineCount: lineCount,
				},
			}),
		});

		const data = await response.json();

		if (response.status == 500) {
			const parser = new DOMParser();
			const errDocument = parser.parseFromString(data.errHTML, "text/html");
			document.replaceChild(
				errDocument.documentElement,
				document.documentElement
			);
			return;
		}

		const addr = data.snippet.addr;

		// NOTE: we can use window.history.pushState to avoid refreshing page
		// but it won't update the go template
		window.location.replace(addr);
	});
}

const lineNumbers = document.querySelector(".main__numbers");
const lineNumber = document.querySelector(".main__number");

if (snippetText != undefined) {
	snippetText.addEventListener("input", function () {
		lineNumbers.innerHTML = "";
		var textLines = snippetText.value.split("\n");
		for (let i = 0; i < textLines.length; i++) {
			const numberElement = lineNumber.cloneNode(true);
			numberElement.innerText = i + 1;
			lineNumbers.appendChild(numberElement);
		}
		lineCount = textLines.length;
	});

	snippetText.addEventListener("scroll", function () {
		lineNumbers.scrollTo(0, snippetText.scrollTop);
	});
}
