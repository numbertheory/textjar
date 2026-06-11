import { createEditor, ThemePreset } from "@notectl/core";
import { createFullPreset } from "@notectl/core/presets";
import { STARTER_FONTS } from "@notectl/core/fonts";
import { formatDistanceToNow, parseISO } from "date-fns";

// Polyfill for crypto.randomUUID in non-secure contexts (e.g. http access via IP)
if (
  typeof window !== "undefined" &&
  window.crypto &&
  !window.crypto.randomUUID
) {
  window.crypto.randomUUID = function () {
    return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) =>
      (
        c ^
        (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))
      ).toString(16),
    );
  };
}

async function initEditor() {
  const container = document.getElementById("editor-container");
  if (!container) return;

  const editor = await createEditor({
    ...createFullPreset({
      font: { fonts: STARTER_FONTS },
      video: false,
    }),
    theme: ThemePreset.Light,
    placeholder: "Enter your text here...",
    autofocus: true,
  });

  const initialContentDiv = document.getElementById("initial-content");
  if (initialContentDiv) {
    await editor.setContentHTML(initialContentDiv.innerHTML);
  }

  container.appendChild(editor);

  const form = document.querySelector("#paste-form");
  if (form) {
    form.onsubmit = async function (e) {
      e.preventDefault();
      const contentInput = document.querySelector("#content-input");

      if (editor.isEmpty()) {
        alert("Please enter some content before saving.");
        return false;
      }

      contentInput.value = await editor.getContentHTML();
      form.submit();
    };
  }
}

function initRelativeDates() {
  const dateElements = document.querySelectorAll(".relative-date");
  dateElements.forEach((el) => {
    const isoDate = el.getAttribute("data-date");
    if (isoDate) {
      try {
        const date = parseISO(isoDate);
        el.textContent = formatDistanceToNow(date, { addSuffix: true });
      } catch (e) {
        console.error("Error parsing date:", e);
      }
    }
  });
}

initEditor().catch(console.error);
initRelativeDates();
