(function () {
  var savedFocus = null;
  var optimisticState = null;

  var ON_CLASSES = ["bg-green-50", "text-green-700"];
  var OFF_CLASSES = ["bg-slate-100", "text-slate-600"];

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
    optimisticState = { el: el, wasOn: wasOn };
    if (wasOn) {
      setClasses(el, OFF_CLASSES, ON_CLASSES);
    } else {
      setClasses(el, ON_CLASSES, OFF_CLASSES);
    }
  }

  function rollbackOptimistic() {
    if (!optimisticState) return;
    var el = optimisticState.el;
    if (!document.body.contains(el)) {
      optimisticState = null;
      return;
    }
    if (optimisticState.wasOn) {
      setClasses(el, ON_CLASSES, OFF_CLASSES);
    } else {
      setClasses(el, OFF_CLASSES, ON_CLASSES);
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
    if (target && target.getAttribute("data-cais-optimistic") === "toggle") {
      optimisticToggle(target);
      target.setAttribute("aria-busy", "true");
    }
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
    if (
      savedFocus &&
      typeof savedFocus.focus === "function" &&
      document.body.contains(savedFocus)
    ) {
      savedFocus.focus();
    }
    savedFocus = null;
  });

  document.body.addEventListener("htmx:beforeSwap", function (evt) {
    if (!document.startViewTransition) return;
    var elt = evt.detail.requestConfig && evt.detail.requestConfig.elt;
    if (!elt || !elt.closest("[data-cais-view-transition]")) return;
    evt.detail.shouldSwap = false;
    document.startViewTransition(function () {
      evt.detail.swap();
    });
  });
})();
