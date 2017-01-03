var BookSummaryModel = function(json) {
  var brm = {}, br = {};
  if (json) {
    br = JSON.parse(json);
  } 
  
  brm.uid = m.prop(br.uid || "");
  brm.title = m.prop(br.title || "");
  brm.bookAuthor = m.prop(br.book_author || "");
  brm.bookURL = m.prop(br.book_url || "");
  brm.overviewHTML = m.prop(br.html || "");
  brm.delta = m.prop(br.delta || "");
  brm.isOngoing = m.prop(br.is_ongoing || false);
  brm.chapters = m.prop(br.chapters || []);

  brm._json = function() {
    return {
      uid: brm.uid(),
      title: brm.title(),
      book_author: brm.bookAuthor(),
      book_url: brm.bookURL(),
      html: brm.overviewHTML(),
      delta: brm.delta(),
      is_ongoing: brm.isOngoing(),
      chapters: brm.chapters(),
    };
  };

  var _saver = function() {
    return m.request({
      method: 'PUT',
      url: '/summaries/' + brm.uid(),
      data: brm._json(),
    });
  };
  brm.saver = _saver;

  brm.save = function() {
    _saver().then(function(response) {
      console.log(response);
    });
  };

  var _deleter = function() {
    return m.request({
      method: 'DELETE',
      url: '/summaries/' + brm.uid(),
    });
  };
  brm.deleter = _deleter;

  brm.chapterList = function() {
    var res = ""; var chapters = brm.chapters();
    for (var i = 0; i < chapters.length; i++) {
      res += chapters[i].heading;
      if (i < chapters.length-1) { res += ", "; }
    }
    return res;
  };

  return brm;
};

var BookSummaryDetailsPopupViewModel = (function() {
  var vm = {};
  vm.isShowPopup = m.prop(false);
  vm._bookSummaryModel = {};
  vm.isCreateMode = m.prop(true);

  vm.openPopup = function(bookSummaryModel) {
    vm._bookSummaryModel = bookSummaryModel || BookSummaryModel();
    if (bookSummaryModel) {
      vm.isCreateMode = m.prop(false);
    }
    vm.isShowPopup(true);
    m.redraw();
  };

  vm.closePopup = function() {
    setTimeout(function() {
      vm.isShowPopup(false);
      m.redraw();
    }, 50);
  };
  
  vm.save = function() {
    vm._bookSummaryModel.save();
    vm.closePopup();
    if (!vm.isCreateMode()) {
      setTimeout(function() {
        window.location.reload(true);
      }, 100);
    }
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
                 m(".row", [
                   m(".modal-header.small-12.columns", m("h2", "Enter book details:")),
                   m("form", {role: "form", action: "/summaries", method: "post"}, [
                     m(".medium-6.small-12.columns", [
                       m("label", "Title",
                        m("input", {type: "text", placeholder: "Title", name: "title", value: vm._bookSummaryModel.title(), oninput: m.withAttr("value", vm._bookSummaryModel.title)})),
                       m("label", "Author",
                         m("input", {type: "text", placeholder: "Author", name: "author", value: vm._bookSummaryModel.bookAuthor(), oninput: m.withAttr("value", vm._bookSummaryModel.bookAuthor)})),
                       m("label", "Amazon URL", 
                         m("input", {type: "text", placeholder: "Amazon URL", name: "url", value: vm._bookSummaryModel.bookURL(), oninput: m.withAttr("value", vm._bookSummaryModel.bookURL)})),
                     ]),
                     m(".medium-6.small-12.columns", [
                      m("label", "Chapters", 
                       m("textarea.chapterbox", (function() { 
                         var a = {name: "chapters", 
                           placeholder: "Chapter list, separated by commas"}; 
                         if (!vm.isCreateMode()) { a.readonly = "true"; }
                         return a;
                       })(), vm._bookSummaryModel.chapterList())),
                       vm.isCreateMode() ? m("input.button.float-right.success", {type: "submit", value: "Go go go!"}) : m(".button.float-right.success", {onclick: vm.save.bind(this)}, "Update!"),
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
  BookSummaryDetailsPopupViewModel.openPopup();
};
