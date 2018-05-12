import _ from 'lodash'

const forwardPropRe = /^(id|className|title|data-.*)$/

export default function forwardProps(self, whitelist) {
  whitelist = whitelist || {}
  // TODO: check for unexpected props being swallowed
  return _.pickBy(self.props, (v, k) => forwardPropRe.test(k) || whitelist[k])
}
