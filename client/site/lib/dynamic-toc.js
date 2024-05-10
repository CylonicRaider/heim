const toc = document.querySelector('.toc')
const inlineAnchor = document.querySelector('#marker-toc')
const sidebar = document.querySelector('.sidebar')
const sidebarContainer = document.querySelector('.sidebar-container')
const selectableEntries = []
let isInSidebar = false
let curSelected = null

function parseFloatOrZero(text) {
  return text ? parseFloat(text) : 0
}

function scrollToShow(node) {
  const parent = node.offsetParent
  if (parent === null) return
  const nodeRect = node.getBoundingClientRect()
  const nodeStyle = getComputedStyle(node)
  const parentRect = parent.getBoundingClientRect()
  const nodeTop = nodeRect.top - parseFloatOrZero(nodeStyle.scrollMarginTop)
  const nodeBottom = nodeRect.bottom + parseFloatOrZero(nodeStyle.scrollMarginBottom)
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
    const newLink = selectableEntries[index].link
    newLink.classList.add('selected')
    if (isInSidebar) {
      scrollToShow(newLink)
    }
  }
}

function updateTocCurrent() {
  if (!selectableEntries.length || !isInSidebar) return

  function testLocation(node) {
    const rect = node.getBoundingClientRect()
    const styles = getComputedStyle(node)
    const top = rect.top - parseFloatOrZero(styles.marginTop)
    const bottom = rect.bottom + parseFloatOrZero(styles.marginBottom)
    if (top > 0) return -1
    if (bottom < 0) return 1
    return 0
  }

  let s = curSelected === null ? 0 : curSelected
  for (;;) {
    const verdict = testLocation(selectableEntries[s].heading)
    if (verdict === 0) break
    if (verdict === -1 && s <= 0) break
    if (verdict === 1 && s >= selectableEntries.length - 1) break
    if (verdict === 1 && testLocation(selectableEntries[s + 1].heading) < 0) break
    s += verdict
  }
  setSelected(s)
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

if (toc && inlineAnchor && sidebar) {
  Array.prototype.forEach.call(toc.querySelectorAll('a'), (link) => {
    if (link.classList.contains('header-anchor') || !/^#[\w-]+/.test(link.hash)) return
    const heading = document.getElementById(link.hash.substring(1))
    selectableEntries.push({
      link: link,
      heading: heading,
    })
  })

  window.addEventListener('resize', updateTocLocation)
  document.addEventListener('scroll', updateTocCurrent)
  updateTocLocation()
  updateTocCurrent()
}
