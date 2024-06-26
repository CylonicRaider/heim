import assert from 'assert'
import Immutable from 'immutable'

import mention from '../lib/heim/mention'

describe('mention', () => {
  describe('findSubseq', () => {
    function assertFinding(n, p, start) {
      assert.deepEqual(mention.findSubseq(n, p), start)
    }

    it('handles empty and exact matches', () => {
      assertFinding('', '', -1)
      assertFinding('name', '', -1)
      assertFinding('', 'cruft', -1)
      assertFinding('name', 'name', 0)
    })
    it('rejects extra on end of partial', () => {
      assertFinding('name', 'namex', -1)
    })
    it('does not case-fold', () => {
      assertFinding('name', 'NAME', -1)
      assertFinding('NAME', 'name', -1)
      assertFinding('NAME', 'NAME', 0)
    })
    it('yields correct start in corner cases', () => {
      assertFinding('a', 'a', 0)
      assertFinding('name', 'n', 0)
      assertFinding('name', 'nm', 0)
      assertFinding('name', 'e', 3)
    })
    it('finds earliest subseq', () => {
      assertFinding('aaaa', 'a', 0)
      assertFinding('namename', 'am', 1)
      assertFinding('ABxxBxCxxD', 'BCD', 1)
    })
    // This is a possible index bug.
    it('rejects extra chars going off the end', () => {
      assertFinding('name', 'es', -1)
    })
    it('respects order', () => {
      assertFinding('name', 'eman', -1)
    })
    it('respects count (always advance)', () => {
      assertFinding('name', 'nnm', -1)
    })
  })

  // These tests are more perfunctory since scoreMatch is the real
  // deal; this is just a supporting fn.
  describe('matchPinningCase', () => {
    function check(name, prefix, match) {
      assert.deepEqual(mention.matchPinningCase(name, prefix), match)
    }
    it('matches without case-folding', () => {
      check('TimMc', 'timmc', null)
      check('timmc', 'TimMc', null)
      check('TimMc', 'TimMc', [-1, -1, 0])
    })
    it('prefers contiguous match over earlier subsequence', () => {
      check('timmc', 'mc', [-1, 0, 3])
      check('timxc', 'mc', [0, 0, 2])
    })
  })

  describe('scoreMatch', () => {
    function scoreEqual(name, prefix, score) {
      assert.deepEqual(mention.scoreMatch(name, prefix), score)
    }

    it('scores entirely bad matches as null', () => {
      scoreEqual('', 'something', null)
      scoreEqual('1', '2', null)
    })
    it('empty prefix doesn\'t match', () => {
      scoreEqual('', '', null)
      scoreEqual('something', '', null)
    })
    it('produces lower scores for better matches', () => {
      // Exact match scores same as prefix for now.
      scoreEqual('TimMc', 'TimMc', [-1, -1, -1, 0])
      scoreEqual('TimMc', 'timmc', [-1, -1, 0, 0])
      scoreEqual('TimMc', 'Tim', [-1, -1, -1, 0])
      scoreEqual('TimMc', 'tim', [-1, -1, 0, 0])
      scoreEqual('TimMc', 'imM', [-1, 0, -1, 1])
      scoreEqual('TimMc', 'imm', [-1, 0, 0, 1])
      scoreEqual('TimMc', 'mc', [-1, 0, 0, 3])
      scoreEqual('TimMc', 'TM', [0, 0, -1, 0])
      scoreEqual('TimMc', 'tm', [0, 0, 0, 0])
      scoreEqual('TimMc', 'no', null)
    })
  })

  describe('rankCompletions', () => {
    function assertRanking(names, part, outNames) {
      const actual = mention
        .rankCompletions(Immutable.Seq(names), part)
        .toArray()
      assert.deepEqual(actual, outNames)
    }
    const users = ['chromakode', 'logan', 'mac', 'Max 2', 'TimMc']

    it('removes entries that don\'t match at all', () => {
      assertRanking(users, 'g', ['logan'])
    })
    it('puts prefix over infix and tie-breaks with case', () => {
      // Note that this demonstrates that a case-insensitive prefix
      // match is superior to a case-sensitive infix match; otherwise
      // we could treat prefix as a high-scoring infix.
      assertRanking(users, 'M', ['Max2', 'mac', 'TimMc', 'chromakode'])
    })
    it('infix over subseq, even when earlier and case-match', () => {
      assertRanking(users, 'mc', ['TimMc', 'mac'])
    })
    it('sorts asciibetically as tie-breaker, caps first', () => {
      assertRanking(['ax', 'Ax', 'Zx', 'zx'], 'x', ['Ax', 'ax', 'Zx', 'zx'])
    })
    it('ranks earlier infix over later', () => {
      // use a vs b to check for interference from fallback sorting
      assertRanking(['aXX', 'bbXX'], 'XX', ['aXX', 'bbXX'])
      assertRanking(['aaXX', 'bXX'], 'XX', ['bXX', 'aaXX'])
    })
    it('ranks earlier subseq over later', () => {
      assertRanking(['xAxB', 'xxAxB'], 'ab', ['xAxB', 'xxAxB'])
      assertRanking(['xBxA', 'xxBxA'], 'ba', ['xBxA', 'xxBxA'])
    })
    it('strips spaces from names', () => {
      assertRanking([' Max 2 '], 'x2', ['Max2'])
    })
  })
})
