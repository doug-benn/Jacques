let wasm = null;

async function init() {
  try {
    const go = new Go();
    const result = await WebAssembly.instantiateStreaming(
      fetch("main.wasm"),
      go.importObject,
    );
    wasm = result.instance;
    go.run(wasm);
    document.getElementById("loading").classList.add("hidden");
  } catch (err) {
    document.getElementById("loading").textContent =
      "Error loading WASM: " + err.message;
    console.error(err);
  }
}

function updateStatus(text) {
  const statusEl = document.getElementById("status");
  statusEl.textContent = text;
  statusEl.className = "status";
  if (text === "Voilà. It is clean.") {
    statusEl.classList.add("done");
  } else if (text === "Error") {
    statusEl.classList.add("error");
  }
}

function showToast(message, type = "") {
  const toast = document.getElementById("toast");
  toast.textContent = message;
  toast.className = "toast show " + type;
  setTimeout(() => {
    toast.className = "toast";
  }, 2000);
}

function clean() {
  if (!window.ready || !window.ready()) {
    updateStatus("Not ready...");
    return;
  }

  const input = document.getElementById("input");
  const output = document.getElementById("output");
  const experimental = document.getElementById("experimental").checked;

  if (!input.value.trim()) {
    output.value = "";
    output.classList.remove("error");
    document.getElementById("status").textContent = "";
    document.getElementById("status").className = "status";
    updateOutputState();
    return;
  }

  updateStatus("Decontaminating...");

  requestAnimationFrame(() => {
    try {
      const result = window.cleanSchema(input.value, experimental);
      output.value = result;
      output.classList.remove("error");
      updateStatus("Voilà. It is clean.");
      updateOutputState();
    } catch (err) {
      output.value = "Error: " + err.message;
      output.classList.add("error");
      updateStatus("Error");
      showToast("Error processing schema", "error");
      updateOutputState();
      console.error(err);
    }
  });
}

function copyOutput() {
  const output = document.getElementById("output");
  if (!output.value) {
    showToast("Nothing to copy", "error");
    return;
  }
  navigator.clipboard
    .writeText(output.value)
    .then(() => {
      showToast("Copied to clipboard!", "success");
    })
    .catch((err) => {
      showToast("Failed to copy", "error");
      console.error(err);
    });
}

function downloadOutput() {
  const output = document.getElementById("output").value;
  if (!output) {
    showToast("Nothing to download", "error");
    return;
  }
  const blob = new Blob([output], { type: "text/plain" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "cleaned_schema.sql";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
  showToast("Downloading schema...", "success");
}

function clearAll() {
  document.getElementById("input").value = "";
  document.getElementById("output").value = "";
  document.getElementById("output").classList.remove("error");
  updateInputState();
  updateOutputState();
  updateStatus("");
}

let debounceTimer;
const inputEl = document.getElementById("input");
const outputEl = document.getElementById("output");
const inputPanel = document.getElementById("inputPanel");
const outputPanel = document.getElementById("outputPanel");
const experimentalEl = document.getElementById("experimental");
const copyBtn = document.getElementById("copyBtn");
const downloadBtn = document.getElementById("downloadBtn");
const clearBtn = document.getElementById("clearBtn");
const scrollBtn = document.getElementById("scrollBtn");
const aboutLink = document.getElementById("aboutLink");
const fileInput = document.getElementById("fileInput");
const dropZone = document.getElementById("dropZone");

function updateInputState() {
  if (inputEl.value.trim()) {
    inputPanel.classList.add("has-content");
  } else {
    inputPanel.classList.remove("has-content");
  }
}

function updateOutputState() {
  const hasContent = !!outputEl.value.trim();
  if (hasContent) {
    outputPanel.classList.add("has-content");
    copyBtn.disabled = false;
    downloadBtn.disabled = false;
  } else {
    outputPanel.classList.remove("has-content");
    copyBtn.disabled = true;
    downloadBtn.disabled = true;
  }
}

inputEl.addEventListener("input", () => {
  updateInputState();
  updateStatus("Decontaminating...");
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(clean, 150);
  hideScrollIndicator();
});

experimentalEl.addEventListener("change", () => {
  updateStatus("Decontaminating...");
  clean();
});

copyBtn.addEventListener("click", copyOutput);
downloadBtn.addEventListener("click", downloadOutput);
clearBtn.addEventListener("click", clearAll);

const scrollToAbout = (e) => {
  if (e) e.preventDefault();
  document.getElementById("about").scrollIntoView({ behavior: "smooth" });
};

scrollBtn.addEventListener("click", scrollToAbout);
aboutLink.addEventListener("click", scrollToAbout);

// File Selection Support
inputPanel.addEventListener("click", (e) => {
  // Trigger ONLY if clicking the "Choose File" link when empty
  const isEmpty = !inputEl.value.trim();
  const clickedLink = e.target.classList.contains("link-style");

  if (isEmpty && clickedLink) {
    fileInput.click();
  }
});

fileInput.addEventListener("change", (e) => {
  const file = e.target.files[0];
  if (file) {
    handleFile(file);
  }
});

function handleFile(file) {
  const reader = new FileReader();
  reader.onload = (event) => {
    inputEl.value = event.target.result;
    updateInputState();
    clean();
    hideScrollIndicator();
  };
  reader.readAsText(file);
}

// Drag and Drop Support
inputPanel.addEventListener("dragover", (e) => {
  e.preventDefault();
  if (!inputEl.value.trim()) {
    inputPanel.classList.add("drag-over");
    dropZone.classList.add("drag-over");
  }
});

inputPanel.addEventListener("dragleave", () => {
  inputPanel.classList.remove("drag-over");
  dropZone.classList.remove("drag-over");
});

inputPanel.addEventListener("drop", (e) => {
  e.preventDefault();
  inputPanel.classList.remove("drag-over");
  dropZone.classList.remove("drag-over");

  const file = e.dataTransfer.files[0];
  if (file) {
    handleFile(file);
  }
});

// Hide the scroll hint permanently once the user scrolls down
const hideScrollIndicator = () => {
  setTimeout(() => {
    scrollBtn.classList.add("hidden");
    observer.disconnect();
    window.removeEventListener("scroll", scrollFallback);
  }, 2000);
};

const observer = new IntersectionObserver(
  (entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        hideScrollIndicator();
      }
    });
  },
  { threshold: 0.2 },
);

observer.observe(document.getElementById("about"));

// Fallback scroll listener
const scrollFallback = () => {
  if (window.scrollY > 100) {
    hideScrollIndicator();
  }
};
window.addEventListener("scroll", scrollFallback);

// Keyboard shortcuts
document.addEventListener("keydown", (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
    clean();
  }
  if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === "C") {
    copyOutput();
  }
});

// Resize handle
const resizeHandle = document.querySelector(".resize-handle");
const container = document.querySelector(".container");
let isResizing = false;

resizeHandle.addEventListener("mousedown", (e) => {
  isResizing = true;
  container.classList.add("resizing");
});

document.addEventListener("mousemove", (e) => {
  if (!isResizing) return;

  const containerRect = container.getBoundingClientRect();
  const handleWidth = resizeHandle.offsetWidth;

  const style = window.getComputedStyle(container);
  const paddingLeft = parseInt(style.paddingLeft) || 0;
  const paddingRight = parseInt(style.paddingRight) || 0;
  const paddingTotal = paddingLeft + paddingRight;

  const availableWidth = containerRect.width - paddingTotal - handleWidth;
  const contentLeft = containerRect.left + paddingLeft;
  const newLeftWidth = e.clientX - contentLeft - handleWidth / 2;

  let percentage = (newLeftWidth / availableWidth) * 100;
  percentage = Math.max(10, Math.min(90, percentage));

  container.style.setProperty('--panel-width', percentage + '%');
});

document.addEventListener("mouseup", () => {
  isResizing = false;
  container.classList.remove("resizing");
});

init();
