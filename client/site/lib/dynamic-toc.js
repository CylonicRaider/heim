/* eslint-disable prefer-arrow-callback, no-var */
(function initDynamicToc() {
  var toc = document.querySelector('.toc')
  var inlineAnchor = document.querySelector('#marker-toc')
  var sidebar = document.querySelector('.sidebar')
  var sidebarContainer = document.querySelector('.sidebar-container')
  var isInSidebar = false
  var selectableEntries = []
  var curSelected = null

  function parseFloatOrZero(text) {
    return text ? parseFloat(text) : 0
  }

  function scrollToShow(node) {
    var parent = node.offsetParent
    if (parent === null) return
    var nodeRect = node.getBoundingClientRect()
    var nodeStyle = getComputedStyle(node)
    var parentRect = parent.getBoundingClientRect()
    var nodeTop = nodeRect.top - parseFloatOrZero(nodeStyle.scrollMarginTop)
    var nodeBottom = nodeRect.bottom + parseFloatOrZero(nodeStyle.scrollMarginBottom)
    if (nodeTop < parentRect.top) {
      parent.scrollTop -= parentRect.top - nodeTop
    } else if (nodeBottom > parentRect.bottom) {
      parent.scrollTop += nodeBottom - parentRect.bottom
    }
  }

  function setSelected(index) {
    if (index === curSelected) {
      return
    }
    if (curSelected !== null) {
      selectableEntries[curSelected].link.classList.remove('selected')
    }
    curSelected = index
    if (index !== null) {
      var newLink = selectableEntries[index].link
      newLink.classList.add('selected')
      if (isInSidebar) {
        scrollToShow(newLink)
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
      var top = rect.top - parseFloatOrZero(styles.marginTop)
      var bottom = rect.bottom + parseFloatOrZero(styles.marginBottom)
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
