(function () {
  var savedFocus = null;
  var optimisticState = null;

  var ON_CLASSES = ["bg-green-50", "text-green-700"];
  var OFF_CLASSES = ["bg-slate-100", "text-slate-600"];

  var NAV_ON = ["bg-slate-900", "text-white", "shadow-2xs"];
  var NAV_OFF = ["text-slate-600", "hover:text-slate-900", "hover:bg-slate-100"];

  function hasClasses(el, classes) {
    return classes.every(function (c) {
      return el.classList.contains(c);
    });
  }

  function setClasses(el, add, remove) {
    remove.forEach(function (c) {
      el.classList.remove(c);
    });
    add.forEach(function (c) {
      el.classList.add(c);
    });
  }

  function optimisticTarget(elt) {
    if (!elt) return null;
    if (elt.matches("[data-cais-optimistic]")) return elt;
    return elt.closest("[data-cais-optimistic]");
  }

  function optimisticToggle(el) {
    var wasOn = hasClasses(el, ON_CLASSES);
    optimisticState = { el: el, wasOn: wasOn, mode: "toggle" };
    if (wasOn) {
      setClasses(el, OFF_CLASSES, ON_CLASSES);
    } else {
      setClasses(el, ON_CLASSES, OFF_CLASSES);
    }
  }

  function optimisticCount(el) {
    var countEl = el.querySelector("[data-cais-count]") || el;
    var raw = countEl.textContent.trim();
    var n = parseInt(raw, 10);
    if (isNaN(n)) n = 0;
    optimisticState = { el: el, mode: "count", countEl: countEl, prev: raw };
    countEl.textContent = String(n + 1);
  }

  function optimisticRemove(el) {
    optimisticState = { el: el, mode: "remove", hadOpacity: el.classList.contains("opacity-0") };
    el.classList.add("opacity-0", "transition-opacity", "duration-150");
  }

  function rollbackOptimistic() {
    if (!optimisticState) return;
    var el = optimisticState.el;
    if (!document.body.contains(el)) {
      optimisticState = null;
      return;
    }
    switch (optimisticState.mode) {
      case "count":
        optimisticState.countEl.textContent = optimisticState.prev;
        break;
      case "remove":
        if (!optimisticState.hadOpacity) {
          el.classList.remove("opacity-0", "transition-opacity", "duration-150");
        }
        break;
      default:
        if (optimisticState.wasOn) {
          setClasses(el, ON_CLASSES, OFF_CLASSES);
        } else {
          setClasses(el, OFF_CLASSES, ON_CLASSES);
        }
    }
    optimisticState = null;
  }

  document.body.addEventListener("htmx:configRequest", function (evt) {
    var meta = document.querySelector('meta[name="csrf-token"]');
    if (meta && meta.content) {
      evt.detail.headers["X-CSRF-Token"] = meta.content;
    }
  });

  document.body.addEventListener("htmx:beforeRequest", function (evt) {
    savedFocus = document.activeElement;
    var target = optimisticTarget(evt.detail.elt);
    if (!target) return;
    var mode = target.getAttribute("data-cais-optimistic");
    if (mode === "toggle") {
      optimisticToggle(target);
    } else if (mode === "count") {
      optimisticCount(target);
    } else if (mode === "remove") {
      optimisticRemove(target);
    } else {
      return;
    }
    target.setAttribute("aria-busy", "true");
  });

  document.body.addEventListener("htmx:responseError", function (evt) {
    rollbackOptimistic();
    var target = optimisticTarget(evt.detail.elt);
    if (target) {
      target.removeAttribute("aria-busy");
    }
  });

  document.body.addEventListener("htmx:afterSettle", function () {
    optimisticState = null;
    document.querySelectorAll("[data-cais-optimistic][aria-busy]").forEach(function (el) {
      el.removeAttribute("aria-busy");
    });
    syncNavTabs();
    dismissExistingToast();
    if (
      savedFocus &&
      typeof savedFocus.focus === "function" &&
      document.body.contains(savedFocus)
    ) {
      savedFocus.focus();
    }
    savedFocus = null;
  });

  document.addEventListener("DOMContentLoaded", function () {
    syncNavTabs();
    dismissExistingToast();
  });

  function syncNavTabs() {
    var nav = document.getElementById("cais-nav");
    if (!nav) return;
    var path = window.location.pathname;
    nav.querySelectorAll("a[data-cais-nav]").forEach(function (a) {
      var href = a.getAttribute("data-cais-nav");
      var active = href === path;
      setClasses(a, active ? NAV_ON : NAV_OFF, active ? NAV_OFF : NAV_ON);
    });
  }

  var toastTimer = null;
  var toastDurationMs = 2000;

  function dismissExistingToast() {
    var host = document.getElementById("cais-toast-host");
    if (!host) return;
    var toast = host.querySelector(".cais-toast-enter");
    if (!toast || !toast.textContent.trim()) return;
    if (toastTimer) {
      clearTimeout(toastTimer);
      toastTimer = null;
    }
    toastTimer = setTimeout(function () {
      host.innerHTML = "";
      toastTimer = null;
    }, toastDurationMs);
  }

  function showToast(message) {
    if (!message) return;
    var host = document.getElementById("cais-toast-host");
    if (!host) return;
    if (toastTimer) {
      clearTimeout(toastTimer);
      toastTimer = null;
    }
    host.innerHTML =
      '<div class="cais-toast-enter fixed top-24 left-1/2 -translate-x-1/2 z-50 bg-slate-900 text-white px-5 py-3 rounded-2xl shadow-xl flex items-center gap-2 border border-slate-700/50" role="status">' +
      '<svg class="w-5 h-5 text-amber-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">' +
      '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />' +
      "</svg>" +
      '<span class="text-xs font-bold"></span></div>';
    host.querySelector("span").textContent = message;
    toastTimer = setTimeout(function () {
      host.innerHTML = "";
      toastTimer = null;
    }, toastDurationMs);
  }

  function parseToastFromTrigger(trigger) {
    if (!trigger) return "";
    try {
      var data = JSON.parse(trigger);
      if (data && data.caisToast) return data.caisToast;
    } catch (e) {
      if (trigger === "caisToast") return "";
    }
    return "";
  }

  function applyTriggerActions(trigger) {
    if (!trigger) return;
    try {
      var data = JSON.parse(trigger);
      if (data && data.caisFocus) {
        var focusEl = document.querySelector(data.caisFocus);
        if (focusEl && typeof focusEl.focus === "function") {
          focusEl.focus();
          if (typeof focusEl.scrollIntoView === "function") {
            focusEl.scrollIntoView({ block: "nearest", behavior: "smooth" });
          }
        }
      }
      if (data && data.caisToast) {
        showToast(data.caisToast);
      }
    } catch (e) {
      if (trigger === "caisToast") return;
    }
  }

  document.body.addEventListener("htmx:afterSwap", function (evt) {
    var xhr = evt.detail.xhr;
    if (!xhr) return;
    applyTriggerActions(xhr.getResponseHeader("HX-Trigger"));
  });

  document.body.addEventListener("click", function (evt) {
    var btn = evt.target.closest("[data-cais-password-toggle]");
    if (!btn) return;
    var wrap = btn.closest(".relative");
    if (!wrap) return;
    var input = wrap.querySelector("input");
    if (!input) return;
    var show = input.type === "password";
    input.type = show ? "text" : "password";
    btn.setAttribute("aria-label", show ? "Hide password" : "Show password");
    var showIcon = btn.querySelector('[data-cais-password-icon="show"]');
    var hideIcon = btn.querySelector('[data-cais-password-icon="hide"]');
    if (showIcon) showIcon.classList.toggle("hidden", show);
    if (hideIcon) hideIcon.classList.toggle("hidden", !show);
  });
})();
