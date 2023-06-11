"use strict";

document.addEventListener("DOMContentLoaded", () => {
	setupColors();
	setupPage();
	if (document.querySelector(".main__input")) {
		setupEditPage();
	}
});

window.addEventListener("keydown", (event) => {
	if (event.key == "s" && event.ctrlKey) {
		event.preventDefault();
		document.querySelector("header__save").click();
	}
});

function setupColors() {
	const storage = window.localStorage;
	const themeButton = document.querySelector(".header__theme");

	if (storage.getItem("light-theme") == "true") {
		document.querySelector("html").classList.toggle("light-theme");
		themeButton.src = "./static/svg/moon.svg";
	}

	themeButton.addEventListener("click", () => {
		if (storage.getItem("light-theme") == "true") {
			themeButton.src = "./static/svg/sun.svg";
		} else {
			themeButton.src = "./static/svg/moon.svg";
		}
	});

	themeButton.addEventListener("click", () => {
		if (storage.getItem("light-theme") == "true") {
			storage.setItem("light-theme", false);
		} else {
			storage.setItem("light-theme", true);
		}

		document.querySelector("html").classList.toggle("light-theme");
	});
}

function setupPage() {
	const snippetAddr = document.querySelector(".header__input");

	document.querySelector(".header__share").addEventListener("click", () => {
		if (snippetAddr.value.length == 0) {
			alert("Please save your code first.");
			return;
		}

		navigator.clipboard.writeText(snippetAddr.value);
	});
}

function setupEditPage() {
	const snippetText = document.querySelector(".main__input");
	const snippetSelect = document.querySelector(".header__select");
	let lineCount = 0;

	document
		.querySelector(".header__save")
		.addEventListener("click", async () => {
			if (snippetText.value.length == 0) {
				alert("Please paste your code.");
				return;
			}

			const response = await fetch("/snippet", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({
					text: snippetText.value,
					lang: snippetSelect.value,
					lineCount: lineCount,
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

			if (data.address == undefined) {
				throw "url adresss should not be undefined";
			}

			// NOTE: we can use window.history.pushState to avoid refreshing page
			// but it won't update the go template
			window.location.replace(data.address);
		});

	const lineNumbers = document.querySelector(".main__numbers");
	const lineNumber = document.querySelector(".main__number");

	snippetText.addEventListener("input", () => {
		lineNumbers.innerHTML = "";
		var textLines = snippetText.value.split("\n");
		for (let i = 0; i < textLines.length; i++) {
			const numberElement = lineNumber.cloneNode(true);
			numberElement.innerText = i + 1;
			lineNumbers.appendChild(numberElement);
		}
		lineCount = textLines.length;
	});

	snippetText.addEventListener("scroll", () => {
		lineNumbers.scrollTo(0, snippetText.scrollTop);
	});
}
