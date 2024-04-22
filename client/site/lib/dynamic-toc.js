(function initDynamicToc() {
  function updateTocLocation() {
    var toc = document.querySelector('.toc')
    if (getComputedStyle(sidebarContainer).display != 'none') {
      sidebar.appendChild(toc)
    } else {
      inlineAnchor.parentNode.insertBefore(toc, inlineAnchor.nextSibling)
    }
  }
  var inlineAnchor = document.querySelector('#marker-toc')
  var sidebar = document.querySelector('.sidebar')
  var sidebarContainer = document.querySelector('.sidebar-container')
  if (!inlineAnchor || !sidebar) return
  window.addEventListener('resize', updateTocLocation)
  updateTocLocation()
})()
