const commands = [
  "terrascale add acme-corp --var project_name=myapp --var subdomain=acme",
  "terrascale add staging --var project_name=myapp --var subdomain=staging",
  "terrascale list --environment=production",
];

const quickstartCommands = {
  bootstrap: {
    copyLabel: "Copy bootstrap command",
    command: `cd your-project
terrascale init`,
  },
  deploy: {
    copyLabel: "Copy deploy command",
  },
  registry: {
    copyLabel: "Copy registry command",
    command: `terrascale list
terrascale inspect staging`,
  },
};

const typedCommand = document.querySelector("#typed-command");
const copyButtons = document.querySelectorAll(".copy-button");
const topbar = document.querySelector(".topbar");
const menuToggle = document.querySelector(".menu-toggle");
const navLinks = document.querySelectorAll(".nav-pill");
const quickstartCommand = document.querySelector("#quickstart-command");
const quickstartCopyButton = document.querySelector('[data-copy-target="quickstart-command"]');
const quickstartSteps = document.querySelectorAll("[data-quickstart-step]");
const quickstartProjectNameInput = document.querySelector("#quickstart-project-name");
const quickstartSubdomainInput = document.querySelector("#quickstart-subdomain");

let quickstartAnimationToken = 0;

function wait(ms) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

async function animateCommands() {
  if (!typedCommand) {
    return;
  }

  let commandIndex = 0;

  while (true) {
    const command = commands[commandIndex];

    for (let i = 0; i <= command.length; i += 1) {
      typedCommand.textContent = command.slice(0, i);
      await wait(24);
    }

    await wait(1600);

    for (let i = command.length; i >= 0; i -= 1) {
      typedCommand.textContent = command.slice(0, i);
      await wait(14);
    }

    await wait(260);
    commandIndex = (commandIndex + 1) % commands.length;
  }
}

async function copyText(button) {
  const targetId = button.dataset.copyTarget;
  const text = targetId ? document.querySelector(`#${targetId}`)?.textContent : button.dataset.copy;
  if (!text) {
    return;
  }

  const originalLabel = button.getAttribute("aria-label") || "Copy command";
  const originalTitle = button.getAttribute("title") || originalLabel;

  try {
    await navigator.clipboard.writeText(text);
    button.dataset.copied = "true";
    button.setAttribute("aria-label", "Copied");
    button.setAttribute("title", "Copied");
  } catch (_error) {
    button.setAttribute("aria-label", "Copy failed");
    button.setAttribute("title", "Copy failed");
  }

  window.setTimeout(() => {
    button.dataset.copied = "false";
    button.setAttribute("aria-label", originalLabel);
    button.setAttribute("title", originalTitle);
  }, 1200);
}

async function animateQuickstartCommand(text) {
  if (!quickstartCommand) {
    return;
  }

  const token = ++quickstartAnimationToken;
  quickstartCommand.textContent = "";

  for (let i = 0; i <= text.length; i += 1) {
    if (token !== quickstartAnimationToken) {
      return;
    }

    quickstartCommand.textContent = text.slice(0, i);
    await wait(10);
  }
}

function getDeployCommand() {
  const deploymentName = (quickstartProjectNameInput?.value || "staging").trim() || "staging";
  const subdomain = (quickstartSubdomainInput?.value || "staging").trim() || "staging";

  return `terrascale add ${deploymentName} \\
  --var project_name=terrascale-${deploymentName} \\
  --var subdomain=${subdomain} \\
  --var environment=${deploymentName}`;
}

function updateQuickstartStep(stepKey) {
  const config = quickstartCommands[stepKey];
  if (!config || !quickstartCopyButton) {
    return;
  }

  quickstartCopyButton.setAttribute("aria-label", config.copyLabel);
  quickstartCopyButton.setAttribute("title", config.copyLabel);

  quickstartSteps.forEach((step) => {
    const isActive = step.dataset.quickstartStep === stepKey;
    step.classList.toggle("is-active", isActive);
    step.setAttribute("aria-selected", String(isActive));
  });

  animateQuickstartCommand(stepKey === "deploy" ? getDeployCommand() : config.command);
}

function wireQuickstartSteps() {
  if (!quickstartSteps.length) {
    return;
  }

  quickstartSteps.forEach((step) => {
    step.addEventListener("click", () => {
      const stepKey = step.dataset.quickstartStep;
      if (!stepKey) {
        return;
      }

      updateQuickstartStep(stepKey);
    });
  });

  updateQuickstartStep("bootstrap");
}

function wireQuickstartInputs() {
  const inputs = [quickstartProjectNameInput, quickstartSubdomainInput].filter(Boolean);
  if (!inputs.length) {
    return;
  }

  inputs.forEach((input) => {
    input.addEventListener("click", (event) => {
      event.stopPropagation();
      updateQuickstartStep("deploy");
    });

    input.addEventListener("input", () => {
      const deployStep = document.querySelector('[data-quickstart-step="deploy"]');
      if (deployStep) {
        quickstartSteps.forEach((step) => {
          const isActive = step === deployStep;
          step.classList.toggle("is-active", isActive);
          step.setAttribute("aria-selected", String(isActive));
        });
      }

      updateQuickstartStep("deploy");
    });
  });
}

function setMenuState(open) {
  if (!topbar || !menuToggle) {
    return;
  }

  topbar.classList.toggle("menu-open", open);
  menuToggle.setAttribute("aria-expanded", String(open));
  menuToggle.setAttribute("aria-label", open ? "Close navigation menu" : "Open navigation menu");
}

function wireMobileMenu() {
  if (!topbar || !menuToggle) {
    return;
  }

  menuToggle.addEventListener("click", () => {
    const isOpen = topbar.classList.contains("menu-open");
    setMenuState(!isOpen);
  });

  navLinks.forEach((link) => {
    link.addEventListener("click", () => {
      setMenuState(false);
    });
  });

  window.addEventListener("resize", () => {
    if (window.innerWidth > 720) {
      setMenuState(false);
    }
  });
}

copyButtons.forEach((button) => {
  button.addEventListener("click", () => {
    copyText(button);
  });
});

wireMobileMenu();
wireQuickstartSteps();
wireQuickstartInputs();
animateCommands();
