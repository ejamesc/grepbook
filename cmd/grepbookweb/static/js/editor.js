var Delta = Quill.import('delta');
var quill = new Quill('#editor', {
  placeholder: 'Start your summary ...',
  theme: 'snow'
});

var brJSON = document.querySelector('#data-bookreview').dataset.bookreviewjson;
var brm = BookSummaryModel(brJSON);

var change = new Delta();
quill.on('text-change', function(delta, source) {
  change = change.compose(delta);
});

var saveText = function() {
  var blah = document.querySelector(".ql-editor").innerHTML;
  brm.overviewHTML(document.querySelector(".ql-editor").innerHTML);
  brm.delta(JSON.stringify(quill.getContents()));
  brm.save();
};

setInterval(function() {
  if (change.length() > 0) {

    // do the save
    change = new Delta();
    saveText();
  }
}, 5*1000);

window.onbeforeunload = function() {
  if (change.length() > 0) {
    return 'There are unsaved changes. Are you sure you want to leave?';
  }
};

document.getElementById("edit-review-button").onclick = function() {
  BookSummaryDetailsPopupViewModel.openPopup(brm);
};
