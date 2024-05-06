/* eslint-disable prefer-arrow-callback, no-var */
(function initDynamicToc() {
  var toc = document.querySelector('.toc')
  var inlineAnchor = document.querySelector('#marker-toc')
  var sidebar = document.querySelector('.sidebar')
  var sidebarContainer = document.querySelector('.sidebar-container')
  var isInSidebar = false
  var selectableEntries = []
  var curSelected = null

  function setSelected(index) {
    if (index === curSelected) {
      return
    }
    if (curSelected !== null) {
      selectableEntries[curSelected].link.classList.remove('selected')
    }
    curSelected = index
    if (index !== null) {
      selectableEntries[index].link.classList.add('selected')
      if (isInSidebar) {
        selectableEntries[index].link.scrollIntoView({block: 'nearest', inline: 'nearest'})
      }
    }
  }

  function updateTocLocation() {
    if (getComputedStyle(sidebarContainer).display !== 'none') {
      sidebar.appendChild(toc)
      isInSidebar = true
      updateTocCurrent()
    } else {
      inlineAnchor.parentNode.insertBefore(toc, inlineAnchor.nextSibling)
      isInSidebar = false
      setSelected(null)
    }
  }

  function updateTocCurrent() {
    if (!selectableEntries.length || !isInSidebar) return

    function testLocation(node) {
      var rect = node.getBoundingClientRect()
      var styles = getComputedStyle(node)
      var top = rect.top - parseFloat(styles.marginTop)
      var bottom = rect.bottom + parseFloat(styles.marginBottom)
      return top > 0 ? -1 : bottom < 0 ? 1 : 0
    }

    var s = curSelected === null ? 0 : curSelected
    for (;;) {
      var verdict = testLocation(selectableEntries[s].heading)
      if (verdict == 0) break
      if (verdict == -1 && s <= 0) break
      if (verdict == 1 && s >= selectableEntries.length - 1) break
      if (verdict == 1 && testLocation(selectableEntries[s + 1].heading) < 0) break
      s += verdict
    }
    setSelected(s)
  }

  if (!toc || !inlineAnchor || !sidebar) return

  Array.prototype.forEach.call(toc.querySelectorAll('a'), function(link) {
    if (link.classList.contains('header-anchor') ||
        !/^#[\w-]+/.test(link.hash))
      return
    var heading = document.getElementById(link.hash.substring(1))
    selectableEntries.push({
      link: link,
      heading: heading,
    })
  })

  window.addEventListener('resize', updateTocLocation)
  document.addEventListener('scroll', updateTocCurrent)
  updateTocLocation()
  updateTocCurrent()
}())
