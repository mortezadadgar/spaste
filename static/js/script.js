"use strict";

document.addEventListener("DOMContentLoaded", () => {
	setupColors();
	setupPage();
	if (document.querySelector(".textarea")) {
		setupEditPage();
	}
});

window.addEventListener("keydown", (event) => {
	if (event.key == "s" && event.ctrlKey) {
		event.preventDefault();
		document.querySelector(".save").click();
	}
});

function setupColors() {
	const storage = window.localStorage;
	const themeButton = document.querySelector(".theme");

	if (storage.getItem("light-theme") == "true") {
		document.querySelector("body").classList.toggle("light-theme");
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

		document.querySelector("body").classList.toggle("light-theme");
	});
}

function setupPage() {
	const pasteAddress = document.querySelector(".address");

	document.querySelector(".share").addEventListener("click", () => {
		if (pasteAddress.value.length == 0) {
			alert("Please save your code first.");
			return;
		}

		navigator.clipboard.writeText(pasteAddress.value);
	});
}

function setupEditPage() {
	const textArea = document.querySelector(".textarea");
	const langSelect = document.querySelector(".select");

	document.querySelector(".save").addEventListener("click", async () => {
		if (textArea.value.length == 0) {
			alert("Please paste your code.");
			return;
		}

		const trimmedText = textArea.value.trim()
		const lineCount = trimmedText.split("\n").length

		const response = await fetch("/paste", {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({
				text: textArea.value,
				lang: langSelect.value,
				lineCount: lineCount,
			}),
		});

		const data = await response.json();

		if (data.address == undefined) {
			throw "can not proceed to empty url address";
		}

		// NOTE: we can use window.history.pushState to avoid refreshing page
		// but it won't update the go template
		window.location.replace(data.address);
	});

	const lineNumbers = document.querySelector(".numbers");
	const lineNumber = document.querySelector(".number");

	textArea.addEventListener("input", () => {
		lineNumbers.innerHTML = "";
		const textLines = textArea.value.split("\n");
		for (let i = 0; i < textLines.length; i++) {
			const numberElement = lineNumber.cloneNode(true);
			numberElement.innerText = i + 1;
			lineNumbers.appendChild(numberElement);
		}
	});

	textArea.addEventListener("scroll", () => {
		lineNumbers.scrollTo(0, textArea.scrollTop);
	});
}
