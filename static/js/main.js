import { createEditor, ThemePreset } from "@notectl/core";
import { createFullPreset } from "@notectl/core/presets";
import { STARTER_FONTS } from "@notectl/core/fonts";

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

initEditor().catch(console.error);
