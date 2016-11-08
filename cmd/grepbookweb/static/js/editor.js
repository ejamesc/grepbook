var Delta = Quill.import('delta');
var quill = new Quill('#editor', {
  placeholder: 'Start your summary ...',
  theme: 'snow'
});

var BookReviewModel = (function() {
  var brJSON = document.querySelector('#data-bookreview').dataset.bookreviewjson;
  var br = JSON.parse(brJSON);
  return br;
})();

var change = new Delta();
quill.on('text-change', function(delta, source) {
  change = change.compose(delta);
});

var saveText = function() {
  BookReviewModel.html = document.querySelector(".ql-editor").innerHTML;
  BookReviewModel.delta = JSON.stringify(quill.getContents());
  var data = BookReviewModel;
  m.request({
    method: 'PUT',
    url: '/summaries/' + BookReviewModel.uid,
    data: data,
  }).then(function(response){
      console.log(response);
    });
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
  BookSummaryDetailsPopupViewModel.openPopup();
};
