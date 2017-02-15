var Delta = Quill.import('delta');

var EditorViewModel = (function() {
  var evm = {};
  var brJSON = document.querySelector('#data-bookreview').dataset.bookreviewjson;
  var _brm = BookSummaryModel(brJSON);
  var quill = null;

  evm.change = new Delta();
  evm.deleter = _brm.deleter;

  // TODO: update the way the html contents are taken?
  function _getText() {
    _brm.overviewHTML(document.querySelector(".ql-editor").innerHTML);
    _brm.delta(JSON.stringify(quill.getContents()));
    evm.change = new Delta(); // we clear it here so we can reuse this in saver+deleter
  }

  evm.chapters = _brm.chapters;

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
    if (!init) {
      quill = new Quill(el, {
        placeholder: 'Start your summary ...',
        theme: 'snow'
      });
      quill.on('text-change', evm.updateDelta);
    }
  };

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
            m("h3", "Overall Book Summary"),
            m("#editor", {config: vm.setup}),
          ]
        )),
      m(".row",
        m(".small-12.medium-10.medium-offset-1.columns",
          (function(){
            var a = [
              m("br"),
              m("h3", "Chapters"),
            ];
            a.push(vm.chapters().map(function(chap, index) {
              return m.component(ChapterEditor, chap);
            }));
            return a;
          })()
        )),
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
    console.log(chap);
    var vm = {};
    vm.editorShown = m.prop(false);
    vm._chap = chap;

    vm.toggleEditor = function() {
      vm.editorShown(!vm.editorShown());
      if (!vm.editorShown()) { cleanupToolbar(this); }
      m.redraw();
    };

    vm.config = function(el, init) {
      if (!init) {
        quill = new Quill(el, {
          placeholder: 'Write your chapter summary ...',
          theme: 'snow'
        });
        quill.on('text-change', vm.updateDelta);
      }
    };

    vm.updateDelta = function() {
      console.log('update delta happening!');
    };

    function cleanupToolbar(el) {
      if (el === null) return;
      var pr = el.parentNode.parentNode;
      var tb = el.parentNode.parentNode.querySelector(".ql-toolbar");
      if (tb === null) return;
      pr.removeChild(tb);
    }

    return vm;
  },
  view: function(vm) {
    return m(".chapter-summary", [
      m("h4.draggable",
        m("span.grey-draggable",
          [m("i.fa.fa-ellipsis-v"), m.trust("&nbsp;&nbsp;")]), 
          m("span", {onclick: vm.toggleEditor}, vm._chap.heading)),
          [(vm.editorShown()) ? m("div", {config: vm.config, id: vm._chap.id}, m.trust(vm._chap.html)) : m.trust(vm._chap.html)]
      ]);
  }
};

m.mount(document.getElementById("summary-placeholder"), Editor);
