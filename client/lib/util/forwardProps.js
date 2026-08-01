import _ from 'lodash'

const forwardPropRe = /^(id|name|className|title|data-.*)$/

export default function forwardProps(self, whitelist) {
  let filterFunc
  if (!whitelist) {
    filterFunc = () => false
  } else if (whitelist instanceof RegExp) {
    filterFunc = whitelist.test.bind(whitelist)
  } else {
    filterFunc = k => whitelist[k]
  }
  // TODO: check for unexpected props being swallowed
  return _.pickBy(self.props, (v, k) => forwardPropRe.test(k) || filterFunc(k))
}
