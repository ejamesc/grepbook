var Delta = Quill.import('delta');

var EditorViewModel = (function() {
  var evm = {};
  var brJSON = document.querySelector('#data-bookreview').dataset.bookreviewjson;
  var _brm = BookSummaryModel(brJSON);
  var _editorEl = null;
  var quill = null;

  evm.change = new Delta();
  evm.deleter = _brm.deleter;
  evm.html = _brm.overviewHTML;
  evm.chapters = _brm._chapters;

  // TODO: update the way the html contents are taken?
  function _getText() {
    _brm.overviewHTML(_editorEl.innerHTML);
    _brm.delta(JSON.stringify(quill.getContents()));
    evm.change = new Delta(); // we clear it here so we can reuse this in saver+deleter
  }

  evm.save = function() {
    _getText();
    _brm.save();
  };

  evm.saver = function() {
    _getText();
    return _brm.saver();
  };

  evm.updateDelta = function(delta, source) {
    evm.change = evm.change.compose(delta);
  };

  evm.openPopup = function() {
    BookSummaryDetailsPopupViewModel.openPopup(_brm);
  };

  evm.updateOngoing = function(ongoing) {
    _brm.isOngoing(ongoing);
    evm.save();
  };

  evm.ongoing = function() {
    return _brm.isOngoing();
  };

  evm.saveButton = function() {
    evm.saver().then(function(r) {
      window.location = "/";
    });
  };

  evm.deleteButton = function() {
    if (confirm("Are you sure you want to delete this review?")) {
      EditorViewModel.deleter().then(function(r) {
        window.location = "/";
      });
    }
  };

  evm.ongoingSwitch = function() {
    if (this.checked) { 
      document.getElementById("ongoing-label").style.display = "block";
    } else {
      document.getElementById("ongoing-label").style.display = "none";
    }
    evm.updateOngoing(this.checked);
  };

  evm.setup = function(el, init) {
    if (init) return;

    el.innerHTML = evm.html();
    quill = new Quill(el, {
      placeholder: 'Start your summary ...',
      theme: 'snow'
    });
    quill.on('text-change', evm.updateDelta);
    _editorEl = el.querySelector(".ql-editor");
  };

  evm.createChapter = function(title) {
    m.request({
      method: "POST",
      url: "/summaries/" + _brm.uid() + "/chapters/",
      data: {heading: title},
    }).then(function(chapJSON) {
      _brm.prependChapter(ChapterModel(chapJSON, _brm));
      m.redraw();
    });
  };

  function getIndexOf(arr, el) {
    if (el === null) return -1;
    for (var i = 0; i < arr.length; i++) {
      if (arr[i].dataset.uid === el.dataset.uid) return i;
    }
    return -1;
  }

  function getModelPosition(id) {
    return _brm._chapters.map(function(c) { return c.id(); }).indexOf(id);
  }

  evm.dragConfig = function(el, init) {
    if (init) return;
    var drake = dragula([el], {
      invalid: function(el, handle) {
        return handle.className.indexOf("draggable") == -1;
      },
    });

    drake.on("drop", function(el, target, source, sibling) {
      var fromIndex = getModelPosition(el.dataset.uid);
      var toIndex = getIndexOf(target.childNodes, el);
      _brm.reorderChapter(fromIndex, toIndex);
    });
  };

  // Autosave
  setInterval(function() {
    if (evm.change.length() > 0) {
      evm.save();
    }
  }, 5*1000);

  window.onbeforeunload = function() {
  if (evm.change.length() > 0) {
    return 'There are unsaved changes. Are you sure you want to leave?';
  }
};

  return evm;
})();

document.getElementById("edit-review-button").onclick = EditorViewModel.openPopup;

var Editor = {
  controller: function() {
    return EditorViewModel;
  },
  view: function(vm) {
    return [
      m(".row", 
        m(".small-12.medium-10.medium-offset-1.columns",
          [
            m("h2", "Overall Book Summary"),
            m("#editor", {config: vm.setup}),
          ]
        )),
      m(".row",
        m(".small-12.medium-10.medium-offset-1.columns", [
              m("br"),
              m("h2", "Chapters"), 
              m.component(CreateChapterWidget, vm),
              m(".sortable", {config: vm.dragConfig},
                vm.chapters.map(function(chap, index) {
                  return m("div", {
                    "data-uid": chap.id(),
                    key: chap.id(),
                    }, m.component(ChapterEditor, chap));
              }))
        ])),
      m(".row",
        m(".small-12.medium-10.medium-offset-1.columns", m("hr"))),
      m(".row", [
        m(".small-12.medium-8.medium-offset-1.columns", 
          [
            m("br"),
            m("input.button.success", {type: "submit", value: "Save", onclick: vm.saveButton}),
            m.trust("&nbsp;"),
            m("button.button.alert", {onclick: vm.deleteButton}, "Delete")
          ]),
        m(".small-12.medium-2.columns.end.text-right", [
          m("label", m("em", "Ongoing?")),
          m(".switch", [
            m("input.switch-input#ongoing-switch", {type: "checkbox", name: "isOngoing", checked: vm.ongoing(), onclick: vm.ongoingSwitch}),
            m("label.switch-paddle", {for: "ongoing-switch"}, [
              m("span.show-for-sr", "Ongoing?"),
              m("span.switch-active", {"aria-hidden": "true"}, "Yes"),
              m("span.switch-inactive", {"aria-hidden": "true"}, "No"),
            ]),
          ]),
        ]),
      ]),
    ];
  },
};

var ChapterEditor = {
  controller: function(chap) {
    var vm = {};
    vm.editorShown = m.prop(false);
    vm.delta = new Delta();
    
    vm._editor = null;
    vm._editorEl = null;
    vm._chap = chap;

    vm.toggleEditor = function() {
      vm.editorShown(!vm.editorShown());
      if (!vm.editorShown()) { cleanupToolbar(); }
      m.redraw();
    };

    vm.config = function(el, init) {
      if (init) return;
      el.innerHTML = vm._chap.html();
      vm._editor = new Quill(el, {
        placeholder: 'Write your chapter summary ...',
        theme: 'snow'
      });
      vm._editor.on('text-change', vm.updateDelta);
      vm._editorEl = el.querySelector(".ql-editor");
    };

    vm.getText = function() {
      return vm._editorEl.innerHTML;
    };

    vm.getDelta = function() {
      return JSON.stringify(vm._editor.getContents());
    };

    vm.updateDelta = function(delta, source) {
      vm.delta = vm.delta.compose(delta);
    };

    vm.delete = function() {
      vm._chap.delete();
    };

    vm.save = function() {
      vm._chap.html(vm.getText());
      vm._chap.delta(vm.getDelta());
      vm._chap.save();
    };

    vm.onSaveClick = function() {
      vm.save();
      vm.toggleEditor();
    };

    vm.onDeleteClick = function() {
      if (confirm("Sure you want to delete this chapter?")) {
        vm._chap.delete();
        vm.toggleEditor();
        m.redraw();
      }
    };

    function cleanupToolbar() {
      vm._editor = null;
      var pr = vm._editorEl.parentNode.parentNode;
      var tb = pr.querySelector(".ql-toolbar");
      pr.removeChild(tb);
      vm._editorEl = null;
    }

    // Autosave
    setInterval(function() {
      if (vm.delta.length() > 0) {
        vm.save();
        vm.delta = new Delta();
      }
    }, 5*1000);

    return vm;
  },
  view: function(vm) {
    return m(".chapter-summary.editable", [
      m("h3.draggable", {onclick: vm.toggleEditor}, 
        m("span.draggable.handle",
          [m("i.fa.fa-ellipsis-v.grey-draggable.draggable"), m.trust("&nbsp;&nbsp;")]), 
          m("span", vm._chap.heading())),
          [(vm.editorShown()) ? m("div", {config: vm.config, id: vm._chap.id()}) : m("span", {onclick: vm.toggleEditor}, m.trust(vm._chap.html()))],
      (vm.editorShown()) ? m(".chapter-footer", [
        m("a.button.primary.small", {onclick: vm.onSaveClick}, m("i.fa.fa-save"), " Save"),
        m.trust("&nbsp;"),
        m("a.button.secondary.small", {onclick: vm.onDeleteClick}, m("i.fa.fa-trash"))]): null,
      ]); 
  }
};

var CreateChapterWidget = {
  controller: function(evm) {
    var vm = {};
    vm.isEditorMode = m.prop(false);
    vm.heading = m.prop("");
    vm.isWarn = m.prop(false);
    
    vm.toggleEditorMode = function() {
      vm.isEditorMode(!vm.isEditorMode());
    };
    
    vm.showWarning = function() {
      vm.isWarn(!vm.isWarn());
      setTimeout(function() {
        vm.isWarn(!vm.isWarn());
        m.redraw();
      }, 2000);
    };

    vm.onCreateClick = function() {
      if (vm.heading() === "") {
        vm.showWarning();
        return;
      }
      evm.createChapter(vm.heading());
      vm.toggleEditorMode();
      vm.heading("");
    };
    return vm;
  },
  view: function(vm) {
    return (vm.isEditorMode()) ?  m("div", [
      m("input[type=text]", {placeholder: "Enter a new chapter title", value: vm.heading(), oninput: m.withAttr("value", vm.heading)}),
      m("a.button.primary.small", {onclick: vm.onCreateClick}, m("i.fa.fa-save"), " Create"),
      m.trust("&nbsp;"),
      m("a. button.secondary.small", {onclick: vm.toggleEditorMode}, m("i.fa.fa-times"), " Cancel"),
      m.trust("&nbsp;"),
      (vm.isWarn()) ? m("span.label.alert", {style: "vertical-align: top;"}, m("i.fa.fa-times-circle-o"), " Chapter title cannot be empty!") : null,
    ]) : m("div.create-chapter-widget", 
        m("a", {onclick: vm.toggleEditorMode}, "+ Add New Chapter"));
  }
};

m.mount(document.getElementById("summary-placeholder"), Editor);
