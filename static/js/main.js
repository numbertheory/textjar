import { createEditor, ThemePreset } from "@notectl/core";
import { createFullPreset } from "@notectl/core/presets";
import { formatDistanceToNow, parseISO } from "date-fns";

const SYSTEM_FONTS = [
  { name: "Arial", family: "Arial, sans-serif", category: "sans-serif" },
  {
    name: "Helvetica",
    family: "Helvetica, sans-serif",
    category: "sans-serif",
  },
  { name: "Georgia", family: "Georgia, serif", category: "serif" },
  {
    name: "Times New Roman",
    family: "'Times New Roman', serif",
    category: "serif",
  },
  {
    name: "Courier New",
    family: "'Courier New', monospace",
    category: "monospace",
  },
];

// Robust Polyfill for crypto.randomUUID in non-secure contexts
(function () {
  const g =
    typeof globalThis !== "undefined"
      ? globalThis
      : typeof window !== "undefined"
        ? window
        : {};
  if (!g.crypto) g.crypto = g.msCrypto || {};

  const crypto = g.crypto;

  // Fallback for getRandomValues if it's also missing in very restrictive contexts
  if (!crypto.getRandomValues) {
    crypto.getRandomValues = function (array) {
      for (let i = 0; i < array.length; i++) {
        array[i] = Math.floor(Math.random() * 256);
      }
      return array;
    };
  }

  if (!crypto.randomUUID) {
    try {
      Object.defineProperty(crypto, "randomUUID", {
        value: function () {
          return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) =>
            (
              c ^
              (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))
            ).toString(16),
          );
        },
        writable: true,
        configurable: true,
      });
    } catch (e) {
      // Fallback to direct assignment if defineProperty fails
      crypto.randomUUID = function () {
        return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(
          /[xy]/g,
          function (c) {
            var r = (Math.random() * 16) | 0,
              v = c == "x" ? r : (r & 0x3) | 0x8;
            return v.toString(16);
          },
        );
      };
    }
  }
})();

async function initEditor() {
  const container = document.getElementById("editor-container");
  if (!container) return;

  const fullPreset = createFullPreset({
    font: { fonts: SYSTEM_FONTS },
  });

  const customToolbar = fullPreset.toolbar
    .map((group) =>
      group.filter((plugin) => plugin.id !== "image" && plugin.id !== "video"),
    )
    .filter((group) => group.length > 0);

  const editor = await createEditor({
    toolbar: customToolbar,
    plugins: fullPreset.plugins,
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
