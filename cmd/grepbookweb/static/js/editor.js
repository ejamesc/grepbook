var Delta = Quill.import('delta');
var quill = new Quill('#summary-placeholder', {
  placeholder: 'Start your summary ...',
  theme: 'snow'
});

var EditorViewModel = (function() {
  var evm = {};
  var brJSON = document.querySelector('#data-bookreview').dataset.bookreviewjson;
  var _brm = BookSummaryModel(brJSON);

  evm.change = new Delta();
  evm.deleter = _brm.deleter;

  // TODO: update the way the html contents are taken?
  function _getText() {
    _brm.overviewHTML(document.querySelector(".ql-editor").innerHTML);
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

quill.on('text-change', EditorViewModel.updateDelta);

document.getElementById("edit-review-button").onclick = EditorViewModel.openPopup;

document.getElementById("save-button").onclick = function() {
  EditorViewModel.saver().then(function(r) {
    window.location = "/";
  });
};

document.getElementById("delete-button").onclick = function() {
  if (confirm("Are you sure you want to delete this review?")) {
    EditorViewModel.deleter().then(function(r) {
      window.location = "/";
    });
  }
};

document.getElementById("ongoing-switch").onclick = function() {
  if (this.checked) { 
    document.getElementById("ongoing-label").style.display = "block";
  } else {
    document.getElementById("ongoing-label").style.display = "none";
  }
  EditorViewModel.updateOngoing(this.checked);
};
