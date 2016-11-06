if (typeof Quill !== 'undefined') {
  var quill = new Quill('#editor', {
    placeholder: 'Start your summary ...',
    theme: 'snow'
  });
}

var BookSummaryDetailsPopupViewModel = (function() {
  var vm = {};
  vm.isShowPopup = m.prop(false);

  vm.openPopup = function() {
    vm.isShowPopup(true);
    m.redraw();
  };

  vm.closePopup = function() {
    setTimeout(function() {
      vm.isShowPopup(false);
      m.redraw();
    }, 50);
  };
  return vm;
})();

var BookSummaryDetailsPopup = {
  controller: function() {
    return BookSummaryDetailsPopupViewModel;
  },
  view: function(vm) {
    if (vm.isShowPopup()) {
      return m(".modal", 
               m(".modal-dialog.modal-white", [
                 m("rows", [
                   m(".modal-header.small-12.columns", m("h2", "Enter book details:")),
                   m("form", {role: "form", action: "/summaries", method: "post"}, [
                     m(".medium-6.small-12.columns", [
                       m("input", {type: "text", placeholder: "Name", name: "name"}),
                       m("input", {type: "text", placeholder: "Author", name: "author"}),
                       m("input", {type: "text", placeholder: "Amazon URL", name: "url"}),
                     ]),
                     m(".medium-6.small-12.columns", [
                       m("textarea.chapterbox", {name: "chapters", placeholder: "Chapter list, separated by commas"}),
                       m("input.button.float-right.success", {type: "submit", value: "Go go go!"}),
                     ]),
                   ]),
                   m("div", {style: "clear:both;"}),
                 ]),
                 m("a.close", {onclick: vm.closePopup}, m.trust("&times;")),
               ]));
    } else {
      return null;
    }
  }
};

m.mount(document.getElementById("modal-placeholder"), BookSummaryDetailsPopup);
document.getElementById("new-review-button").onclick = function() {
  console.log('yay');
  BookSummaryDetailsPopupViewModel.openPopup();
};
