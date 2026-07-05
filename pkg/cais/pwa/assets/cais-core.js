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

  function chatEnabled() {
    return !!document.querySelector("[data-cais-chat]");
  }

  function chatFormFrom(elt) {
    if (!elt) return null;
    if (elt.matches && elt.matches("form[data-cais-chat-form]")) return elt;
    return elt.closest ? elt.closest("form[data-cais-chat-form]") : null;
  }

  function chatInputField(form) {
    return form
      ? form.querySelector(
          "#chat-text, textarea[name='text'], textarea[name='content'], input[name='text'], input[name='content']"
        )
      : null;
  }

  function chatHistoryEl() {
    return document.getElementById("chat-history");
  }

  function chatLiveEl() {
    return document.getElementById("chat-live");
  }

  function chatMessagesEl() {
    return document.getElementById("chat-messages");
  }

  function chatScrollDownBtn() {
    return document.getElementById("chat-scroll-down");
  }

  function formatLocalChatTimestamp(d) {
    var pad = function (n) {
      return String(n).padStart(2, "0");
    };
    return (
      pad(d.getDate()) +
      "/" +
      pad(d.getMonth() + 1) +
      "/" +
      d.getFullYear() +
      " " +
      pad(d.getHours()) +
      ":" +
      pad(d.getMinutes())
    );
  }

  function formatMessageTimes(root) {
    var scope = root || document;
    scope.querySelectorAll("time.cais-msg-time[datetime]").forEach(function (el) {
      var raw = el.getAttribute("datetime");
      if (!raw) return;
      var d = new Date(raw);
      if (isNaN(d.getTime())) return;
      el.textContent = formatLocalChatTimestamp(d);
    });
  }

  function wrapChatBubble(bubble, role, at) {
    var wrap = document.createElement("div");
    wrap.className =
      role === "user"
        ? "cais-msg cais-msg-user max-w-[85%] ml-auto flex flex-col items-end gap-0.5"
        : "cais-msg cais-msg-assistant max-w-[85%] flex flex-col items-start gap-0.5";
    var time = document.createElement("time");
    time.className = "cais-msg-time";
    time.dateTime = at.toISOString();
    time.textContent = formatLocalChatTimestamp(at);
    wrap.appendChild(time);
    wrap.appendChild(bubble);
    return wrap;
  }

  function removeOptimisticUserBubble() {
    document.querySelectorAll("[data-cais-optimistic-user]").forEach(function (el) {
      el.remove();
    });
  }

  function userBubbleText(node) {
    if (!node) return "";
    var bubble = node.querySelector ? node.querySelector(".cais-chat-bubble.user") : null;
    if (bubble) return bubble.textContent.trim();
    return node.textContent.trim();
  }

  function historyHasUserText(history, text) {
    if (!history || !text) return false;
    var nodes = history.querySelectorAll(".cais-msg-user, .cais-chat-bubble.user");
    for (var i = nodes.length - 1; i >= 0; i--) {
      if (userBubbleText(nodes[i]) === text) return true;
    }
    return false;
  }

  function dedupOptimisticUserBubble(history) {
    history = history || chatHistoryEl();
    if (!history) return;
    document.querySelectorAll("[data-cais-optimistic-user]").forEach(function (opt) {
      var text = userBubbleText(opt);
      if (historyHasUserText(history, text)) opt.remove();
    });
  }

  function bindChatEnterSubmit() {
    if (document.documentElement.dataset.caisChatEnterBound === "true") return;
    document.documentElement.dataset.caisChatEnterBound = "true";
    document.addEventListener(
      "keydown",
      function (evt) {
        if (evt.key !== "Enter" || evt.shiftKey || evt.isComposing || evt.defaultPrevented) return;
        var target = evt.target;
        if (!target || !target.matches) return;
        if (!target.matches("textarea, input[type='text']")) return;
        var form = chatFormFrom(target);
        if (!form) return;
        evt.preventDefault();
        if (typeof form.requestSubmit === "function") {
          form.requestSubmit();
        } else {
          form.dispatchEvent(new Event("submit", { cancelable: true, bubbles: true }));
        }
      },
      true
    );
  }

  function appendOptimisticUserBubble(text) {
    if (!chatEnabled()) return;
    var history = chatHistoryEl();
    text = (text || "").trim();
    if (!history || !text) return;
    removeOptimisticUserBubble();
    var bubble = document.createElement("div");
    bubble.className =
      "cais-chat-bubble user cais-user-pending rounded-2xl rounded-br-sm bg-indigo-600 px-4 py-2 text-sm text-white shadow-xs";
    bubble.textContent = text;
    var built = wrapChatBubble(bubble, "user", new Date());
    built.setAttribute("data-cais-optimistic-user", "true");
    history.appendChild(built);
    caisChatScrollBottomSoon();
  }

  var finalizeStreamTimer = null;
  var chatStickToBottom = true;

  function assistantBubbleFromNode(node) {
    if (!node) return null;
    if (
      node.classList &&
      node.classList.contains("cais-chat-bubble") &&
      node.classList.contains("assistant")
    ) {
      return node;
    }
    return node.querySelector ? node.querySelector(".cais-chat-bubble.assistant") : null;
  }

  function historyHasAssistantText(history, text) {
    if (!text) return true;
    var bubbles = history.querySelectorAll(".cais-chat-bubble.assistant");
    for (var i = bubbles.length - 1; i >= 0; i--) {
      if (bubbles[i].textContent.trim() === text) return true;
    }
    return false;
  }

  function streamHasAssistantMessages() {
    var stream = document.getElementById("chat-stream");
    var live = chatLiveEl();
    return !!(
      (stream && stream.querySelector(".cais-chat-bubble.assistant")) ||
      (live && live.querySelector(".cais-chat-bubble.assistant"))
    );
  }

  function appendStreamNode(history, node) {
    var bubble = assistantBubbleFromNode(node);
    if (!bubble) {
      node.remove();
      return;
    }
    var text = bubble.textContent.trim();
    if (!text || historyHasAssistantText(history, text)) {
      node.remove();
      return;
    }
    var built = node.classList.contains("cais-msg")
      ? node
      : wrapChatBubble(bubble, "assistant", new Date());
    node.remove();
    history.appendChild(built);
  }

  function finalizeChatLive() {
    var live = chatLiveEl();
    var history = chatHistoryEl();
    if (!live || !history || !live.firstElementChild) return;
    appendStreamNode(history, live.firstElementChild);
    live.innerHTML = "";
  }

  function finalizeChatStream() {
    if (!chatEnabled()) return;
    finalizeChatLive();
    var stream = document.getElementById("chat-stream");
    var history = chatHistoryEl();
    if (!stream || !history) return;
    Array.from(stream.children).forEach(function (node) {
      appendStreamNode(history, node);
    });
    pruneEmptyChatNodes();
  }

  window.caisFinalizeChatStream = finalizeChatStream;

  function nodeHasVisibleChatContent(node) {
    if (!node) return false;
    if (node.querySelector("img, pre, code, details, .cais-thinking-dots")) return true;
    var bubble =
      node.querySelector(".cais-chat-bubble") ||
      (node.classList.contains("cais-chat-bubble") ? node : null);
    if (bubble && bubble.textContent.trim()) return true;
    return !!node.textContent.trim();
  }

  function pruneEmptyChatNodes() {
    ["#chat-stream", "#chat-live", "#chat-history"].forEach(function (sel) {
      var el = document.querySelector(sel);
      if (!el) return;
      Array.from(el.children).forEach(function (child) {
        if (!nodeHasVisibleChatContent(child)) child.remove();
      });
    });
  }

  function clearChatLive() {
    var live = chatLiveEl();
    if (live) live.innerHTML = "";
  }

  function scheduleFinalizeStream() {
    if (finalizeStreamTimer) clearTimeout(finalizeStreamTimer);
    finalizeStreamTimer = setTimeout(function () {
      finalizeStreamTimer = null;
      finalizeChatStream();
      caisChatScrollBottomSoon();
    }, 700);
  }

  function refreshChatMessages() {
    var el = chatSSEEl();
    if (!el) return;
    var pollURL = el.getAttribute("data-cais-poll-url");
    if (!pollURL || typeof htmx === "undefined") return;
    finalizeChatStream();
    if (streamHasAssistantMessages()) return;
    htmx.ajax("GET", pollURL, { target: "#chat-history", swap: "innerHTML" });
    caisChatScrollBottomSoon();
  }

  function chatIsNearBottom(box, threshold) {
    if (!box) return true;
    threshold = typeof threshold === "number" ? threshold : 96;
    return box.scrollHeight - box.scrollTop - box.clientHeight <= threshold;
  }

  function updateChatScrollDownButton() {
    var box = chatMessagesEl();
    var btn = chatScrollDownBtn();
    if (!box || !btn) return;
    var nearBottom = chatIsNearBottom(box);
    if (nearBottom) {
      chatStickToBottom = true;
    }
    btn.classList.toggle("hidden", nearBottom);
    btn.setAttribute("aria-hidden", nearBottom ? "true" : "false");
  }

  function bindChatScrollDown() {
    var box = chatMessagesEl();
    var btn = chatScrollDownBtn();
    if (!box || !btn || box.dataset.caisScrollBound === "true") return;
    box.dataset.caisScrollBound = "true";
    box.addEventListener(
      "scroll",
      function () {
        chatStickToBottom = chatIsNearBottom(box);
        updateChatScrollDownButton();
      },
      { passive: true }
    );
    btn.addEventListener("click", function () {
      chatStickToBottom = true;
      box.scrollTo({ top: box.scrollHeight, behavior: "smooth" });
      window.setTimeout(function () {
        caisChatScrollBottom();
        updateChatScrollDownButton();
      }, 280);
    });
  }

  function caisChatScrollBottom() {
    if (!chatStickToBottom) {
      updateChatScrollDownButton();
      return;
    }
    var box = chatMessagesEl();
    if (!box) return;
    box.scrollTop = box.scrollHeight;
    var last = chatHistoryEl() && chatHistoryEl().lastElementChild;
    if (last && typeof last.scrollIntoView === "function") {
      last.scrollIntoView({ block: "end", inline: "nearest" });
    }
    updateChatScrollDownButton();
  }

  function caisChatScrollBottomSoon() {
    if (!chatStickToBottom) {
      updateChatScrollDownButton();
      return;
    }
    caisChatScrollBottom();
    requestAnimationFrame(function () {
      caisChatScrollBottom();
      requestAnimationFrame(caisChatScrollBottom);
    });
    [50, 150, 400].forEach(function (ms) {
      setTimeout(caisChatScrollBottom, ms);
    });
  }

  // Improve auto-follow reliability for live streaming (height changes after DOM inserts).
  // Uses ResizeObserver when available so we react to actual bubble growth instead of only timers.
  var chatScrollResizeObserver = null;
  function bindChatAutoScrollResize() {
    var box = chatMessagesEl();
    if (!box || chatScrollResizeObserver || typeof ResizeObserver === "undefined") return;
    chatScrollResizeObserver = new ResizeObserver(function () {
      if (chatStickToBottom) {
        // Defer to next frame to let layout settle (live content + images).
        requestAnimationFrame(caisChatScrollBottom);
      }
    });
    chatScrollResizeObserver.observe(box);
    var live = chatLiveEl();
    if (live) chatScrollResizeObserver.observe(live);
  }

  window.caisChatScrollBottom = caisChatScrollBottom;
  window.caisRemoveOptimisticUserBubble = removeOptimisticUserBubble;

  function bootChatModule() {
    if (!chatEnabled()) return;
    chatStickToBottom = true;
    bindChatScrollDown();
    bindChatAutoScrollResize();
    formatMessageTimes(chatMessagesEl());
    caisChatScrollBottomSoon();
    updateChatScrollDownButton();
  }

  document.body.addEventListener("htmx:sseBeforeMessage", function () {
    hideChatThinking();
    clearChatFallbackTimers();
  });

  // (chat-specific SSE behavior lives in cais-chat.js)

  document.body.addEventListener("htmx:sseClose", function (evt) {
    var el = evt.detail && evt.detail.elt;
    if (!el || !shouldPersistSSE(el)) return;
    if (evt.detail.type === "nodeReplaced") {
      scheduleSSEReconnect();
    }
    if (chatEnabled()) {
      removeOptimisticUserBubble();
      refreshChatMessages();
    }
  });

  document.body.addEventListener("htmx:sseError", function () {
    scheduleSSEReconnect();
    if (chatEnabled()) {
      removeOptimisticUserBubble();
      refreshChatMessages();
    }
  });

  document.body.addEventListener("htmx:beforeRequest", function (evt) {
    savedFocus = document.activeElement;
    var chatForm = chatFormFrom(evt.detail.elt);
    if (chatForm) {
      var field = chatInputField(chatForm);
      if (field) {
        chatForm._caisDraft = field.value;
        if (chatEnabled()) {
          finalizeChatStream();
          if (chatForm.getAttribute("data-cais-chat-optimistic") === "true") {
            appendOptimisticUserBubble(field.value);
          }
          field.value = "";
          caisChatScrollBottomSoon();
        }
      }
      scheduleChatFallback();
    }
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
    var chatForm = chatFormFrom(evt.detail.elt);
    if (chatForm && chatForm._caisDraft !== undefined) {
      var ta = chatInputField(chatForm);
      if (ta) ta.value = chatForm._caisDraft;
      delete chatForm._caisDraft;
      removeOptimisticUserBubble();
      hideChatThinking();
    }
    var target = optimisticTarget(evt.detail.elt);
    if (target) {
      target.removeAttribute("aria-busy");
    }
  });

  var sseReconnectTimer = null;
  var chatFallbackTimers = [];

  function chatSSEEl() {
    return document.getElementById("chat-sse");
  }

  function shouldPersistSSE(el) {
    return el && el.getAttribute("data-cais-sse-persist") === "true";
  }

  function hasActiveSSE(el) {
    if (!el || typeof htmx === "undefined" || !htmx.getInternalData) return false;
    var data = htmx.getInternalData(el);
    return data && data.sseEventSource && data.sseEventSource.readyState !== EventSource.CLOSED;
  }

  function reconnectChatSSE() {
    var el = chatSSEEl();
    if (!el || !shouldPersistSSE(el) || hasActiveSSE(el)) return;
    if (typeof htmx !== "undefined" && htmx.process) {
      htmx.process(el);
    }
  }

  function scheduleSSEReconnect() {
    if (sseReconnectTimer) clearTimeout(sseReconnectTimer);
    sseReconnectTimer = setTimeout(function () {
      sseReconnectTimer = null;
      reconnectChatSSE();
    }, 100);
  }

  function hideChatThinking() {
    var thinking = document.getElementById("chat-thinking");
    if (thinking) thinking.classList.add("hidden");
  }

  function clearChatFallbackTimers() {
    chatFallbackTimers.forEach(function (id) {
      clearTimeout(id);
    });
    chatFallbackTimers = [];
  }

  function scheduleChatFallback() {
    clearChatFallbackTimers();
    var el = chatSSEEl();
    if (!el) return;
    var pollURL = el.getAttribute("data-cais-poll-url");
    if (!pollURL || typeof htmx === "undefined") return;
    var delays = chatEnabled() ? [1500, 3000, 6000, 12000] : [4000, 8000, 15000];
    delays.forEach(function (ms) {
      var id = setTimeout(function () {
        var thinking = document.getElementById("chat-thinking");
        if (!thinking || thinking.classList.contains("hidden")) return;
        if (chatEnabled()) {
          if (streamHasAssistantMessages()) return;
          refreshChatMessages();
        } else {
          htmx.ajax("GET", pollURL, { target: "#chat-history", swap: "innerHTML" });
        }
        hideChatThinking();
      }, ms);
      chatFallbackTimers.push(id);
    });
  }

  document.body.addEventListener("htmx:afterSettle", function () {
    optimisticState = null;
    document.querySelectorAll("[data-cais-optimistic][aria-busy]").forEach(function (el) {
      el.removeAttribute("aria-busy");
    });
    syncNavTabs();
    dismissExistingToast();
    reconnectChatSSE();
    bootChatModule();
    if (
      savedFocus &&
      typeof savedFocus.focus === "function" &&
      document.body.contains(savedFocus)
    ) {
      savedFocus.focus();
    }
    savedFocus = null;
  });

  document.body.addEventListener("htmx:afterSwap", function (evt) {
    var target = evt.detail && evt.detail.target;
    if (!target || !chatEnabled()) return;
    if (target.id === "chat-history") {
      dedupOptimisticUserBubble(target);
      finalizeChatStream();
      pruneEmptyChatNodes();
      formatMessageTimes(target);
      caisChatScrollBottomSoon();
      return;
    }
    if (target.id === "chat-stream" || target.id === "chat-live") {
      formatMessageTimes(target);
      caisChatScrollBottomSoon();
    }
  });

  // chat scroll on sse is in cais-chat.js

  bindChatEnterSubmit();

  document.addEventListener("DOMContentLoaded", function () {
    syncNavTabs();
    dismissExistingToast();
    bootChatModule();
  });

  window.addEventListener("load", bootChatModule);
  window.addEventListener("pageshow", function () {
    bootChatModule();
    reconnectChatSSE();
    if (chatEnabled()) refreshChatMessages();
  });

  document.addEventListener("visibilitychange", function () {
    if (document.visibilityState !== "visible" || !chatEnabled()) return;
    reconnectChatSSE();
    refreshChatMessages();
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

  var openSelectSearch = null;

  function normalizeSelectSearchText(value) {
    return (value || "")
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "");
  }

  function selectSearchLabel(select) {
    var option = select.options[select.selectedIndex];
    if (!option) return "";
    return option.textContent.trim();
  }

  function closeSelectSearchPanel(wrap) {
    if (!wrap) return;
    var panel = wrap.querySelector(".cais-select-search-panel");
    var trigger = wrap.querySelector(".cais-select-search-trigger");
    if (panel) panel.classList.add("hidden");
    if (trigger) trigger.setAttribute("aria-expanded", "false");
    if (openSelectSearch === wrap) openSelectSearch = null;
  }

  function visibleSelectSearchOptions(wrap) {
    var list = wrap.querySelector(".cais-select-search-list");
    if (!list) return [];
    return Array.prototype.filter.call(list.children, function (item) {
      return !item.classList.contains("is-hidden");
    });
  }

  function highlightSelectSearchOption(wrap, optionEl) {
    var list = wrap.querySelector(".cais-select-search-list");
    if (!list) return;
    Array.prototype.forEach.call(list.children, function (item) {
      item.classList.remove("is-highlighted");
    });
    if (optionEl) optionEl.classList.add("is-highlighted");
  }

  function syncSelectSearchOptions(wrap) {
    var select = wrap.querySelector("select[data-cais-select-search]");
    var list = wrap.querySelector(".cais-select-search-list");
    if (!select || !list) return;
    list.innerHTML = "";
    Array.prototype.forEach.call(select.options, function (option) {
      var item = document.createElement("li");
      item.className = "cais-select-search-option";
      item.setAttribute("role", "option");
      item.setAttribute("data-value", option.value);
      item.textContent = option.textContent.trim();
      if (option.selected) item.classList.add("is-selected");
      list.appendChild(item);
    });
  }

  function updateSelectSearchTrigger(wrap) {
    var select = wrap.querySelector("select[data-cais-select-search]");
    var trigger = wrap.querySelector(".cais-select-search-trigger");
    var label = wrap.querySelector(".cais-select-search-label");
    if (!select || !trigger || !label) return;
    label.textContent = selectSearchLabel(select);
    trigger.disabled = !!select.disabled;
  }

  function filterSelectSearchOptions(wrap, query) {
    var list = wrap.querySelector(".cais-select-search-list");
    if (!list) return;
    var needle = normalizeSelectSearchText(query);
    Array.prototype.forEach.call(list.children, function (item) {
      var hay = normalizeSelectSearchText(item.textContent);
      var match = !needle || hay.indexOf(needle) !== -1;
      item.classList.toggle("is-hidden", !match);
      item.classList.remove("is-highlighted");
    });
    var visible = visibleSelectSearchOptions(wrap);
    if (visible.length) highlightSelectSearchOption(wrap, visible[0]);
  }

  function selectSearchValue(wrap, value) {
    var select = wrap.querySelector("select[data-cais-select-search]");
    var list = wrap.querySelector(".cais-select-search-list");
    if (!select || !list) return;
    select.value = value;
    Array.prototype.forEach.call(list.children, function (item) {
      item.classList.toggle("is-selected", item.getAttribute("data-value") === value);
    });
    select.dispatchEvent(new Event("change", { bubbles: true }));
    updateSelectSearchTrigger(wrap);
    closeSelectSearchPanel(wrap);
  }

  function openSelectSearchPanel(wrap) {
    if (!wrap) return;
    if (openSelectSearch && openSelectSearch !== wrap) {
      closeSelectSearchPanel(openSelectSearch);
    }
    var panel = wrap.querySelector(".cais-select-search-panel");
    var trigger = wrap.querySelector(".cais-select-search-trigger");
    var input = wrap.querySelector(".cais-select-search-input");
    if (!panel || !trigger || trigger.disabled) return;
    syncSelectSearchOptions(wrap);
    updateSelectSearchTrigger(wrap);
    panel.classList.remove("hidden");
    trigger.setAttribute("aria-expanded", "true");
    openSelectSearch = wrap;
    if (input) {
      input.value = "";
      filterSelectSearchOptions(wrap, "");
      input.focus();
    }
  }

  function enhanceSelectSearch(select) {
    if (!select || select.getAttribute("data-cais-select-search") === "false") return;
    var wrap = document.createElement("div");
    wrap.className = "cais-select-search";
    var parent = select.parentNode;
    parent.insertBefore(wrap, select);
    wrap.appendChild(select);
    select.classList.add("cais-select-search-native");
    select.setAttribute("tabindex", "-1");
    select.setAttribute("aria-hidden", "true");

    var trigger = document.createElement("button");
    trigger.type = "button";
    trigger.className = "cais-select-search-trigger";
    trigger.setAttribute("aria-haspopup", "listbox");
    trigger.setAttribute("aria-expanded", "false");
    trigger.innerHTML =
      '<span class="cais-select-search-label"></span>' +
      '<svg class="cais-select-search-chevron" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">' +
      '<path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 10.94l3.71-3.71a.75.75 0 111.06 1.06l-4.24 4.24a.75.75 0 01-1.06 0L5.21 8.29a.75.75 0 01.02-1.08z" clip-rule="evenodd" />' +
      "</svg>";

    var panel = document.createElement("div");
    panel.className = "cais-select-search-panel hidden";
    panel.innerHTML =
      '<input type="search" class="cais-select-search-input" placeholder="Search..." autocomplete="off" aria-label="Search options" />' +
      '<ul class="cais-select-search-list" role="listbox"></ul>';

    wrap.appendChild(trigger);
    wrap.appendChild(panel);
    syncSelectSearchOptions(wrap);
    updateSelectSearchTrigger(wrap);

    trigger.addEventListener("click", function () {
      if (panel.classList.contains("hidden")) {
        openSelectSearchPanel(wrap);
      } else {
        closeSelectSearchPanel(wrap);
      }
    });

    var searchInput = panel.querySelector(".cais-select-search-input");
    searchInput.addEventListener("input", function () {
      filterSelectSearchOptions(wrap, searchInput.value);
    });
    searchInput.addEventListener("keydown", function (evt) {
      var visible = visibleSelectSearchOptions(wrap);
      if (!visible.length) return;
      var current = panel.querySelector(".cais-select-search-option.is-highlighted");
      var index = current ? visible.indexOf(current) : -1;
      if (evt.key === "ArrowDown") {
        evt.preventDefault();
        highlightSelectSearchOption(wrap, visible[Math.min(index + 1, visible.length - 1)]);
      } else if (evt.key === "ArrowUp") {
        evt.preventDefault();
        highlightSelectSearchOption(wrap, visible[Math.max(index - 1, 0)]);
      } else if (evt.key === "Enter") {
        evt.preventDefault();
        var pick = panel.querySelector(".cais-select-search-option.is-highlighted") || visible[0];
        if (pick) selectSearchValue(wrap, pick.getAttribute("data-value"));
      } else if (evt.key === "Escape") {
        evt.preventDefault();
        closeSelectSearchPanel(wrap);
        trigger.focus();
      }
    });

    panel.querySelector(".cais-select-search-list").addEventListener("click", function (evt) {
      var option = evt.target.closest(".cais-select-search-option");
      if (!option || option.classList.contains("is-hidden")) return;
      selectSearchValue(wrap, option.getAttribute("data-value"));
    });
  }

  function initSelectSearch(root) {
    var scope = root || document;
    scope
      .querySelectorAll("select[data-cais-select-search]:not([data-cais-select-bound])")
      .forEach(function (select) {
        select.setAttribute("data-cais-select-bound", "true");
        enhanceSelectSearch(select);
      });
  }

  document.body.addEventListener("click", function (evt) {
    if (!openSelectSearch) return;
    if (evt.target.closest(".cais-select-search") === openSelectSearch) return;
    closeSelectSearchPanel(openSelectSearch);
  });

  document.body.addEventListener("htmx:afterSettle", function () {
    initSelectSearch();
  });

  document.addEventListener("DOMContentLoaded", function () {
    initSelectSearch();
  });
})();
