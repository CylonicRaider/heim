import React from 'react'

import { MainPage, FancyLogo, FauxMessage, links } from './common'
import heimURL from '../lib/heimURL'

module.exports = (
  <MainPage title="About Euphoria" className="about">
    <div className="primary">
      <h1>What's Euphoria?</h1>
      <section className="letter">
        <h3>Euphoria is chat for all, in communities that matter.</h3>
        <p>Hello! We're building a platform for chat rooms you care about.</p>
        <p>Social chat rooms have played an important role in each of our lives. They’ve entertained us, offered inspiration and feedback, and given us lasting friends and relationships.</p>
        <p>Our goal is to foster new spaces that can offer this experience to you. Here's a bit more about us. We look forward to learning more about you.</p>
        <div className="end">
          <p className="contact">
            <span className="label">Got a question, or want to learn more?</span>
            <a className="chat-with-us green-button" href={heimURL('/room/welcome/')}>chat with us &raquo;</a>
          </p>
        </div>
      </section>
    </div>
    <section>
      <h2>Our values (in emoji form)</h2>
      <FancyLogo />
      <ul className="values">
        <li className="welcoming">
          <h4>Euphoria is welcoming</h4>
          <p>Friendly to newcomers, easy to join, and safe.</p>
        </li>
        <li className="diverse">
          <h4>Euphoria is diverse</h4>
          <p>We strive for tolerance, fairness, and accessibility.</p>
        </li>
        <li className="meaningful">
          <h4>Euphoria is informal yet meaningful</h4>
          <p>Everyone deserves respect, empathy, and understanding.</p>
        </li>
      </ul>
      <p>For more details, check out <a href={heimURL('/about/values')}>Euphoria's Values</a> statement.</p>
    </section>
    <section>
      <h2>Euphoria is open</h2>
      <h3>We believe that online community platforms should be open source.</h3>
      <p>Our chat server, Heim, is <a href={links.heimSourceRepo}>available on GitHub</a>. Join our development chat in <a href={heimURL('/room/heim')}>&heim</a>.</p>
    </section>
    <section className="who">
      <h2>Who we were</h2>
      <p>This section lists the original creators of Euphoria, who have since left to pursue other projects.</p>
      <div className="old-who">
        <h3>We're a small team who care deeply about online socialization and citizenship.</h3>
        <div className="messages wrap">
          <FauxMessage sender="intortus" message="hi, I'm logan. I'm an erstwhile motorcycle racer and one of euphoria's programmers. you can meet new users with me in &welcome or discuss backend development with me in &heim." />
          <FauxMessage sender="chromakode" message="hey, I'm Max! I live in San Francisco, where I work on Euphoria's user interface and design. you can often find me jamming in &music or working in &heim." />
          <FauxMessage sender="greenie" message="oh hai, I’m Kris. I live with too many cats in the deep forest of Vermont. when I’m not busy getting into moose-caused traffic jams, I do community stuff for Euphoria. I tend to be found linking articles in &space and playing bluegrass tunes in &music." />
          <FauxMessage sender="ezzie" message="hi, I'm ezzie, the office dog!" embed={heimURL('/static/ezzie.jpg')} />
        </div>
      </div>
    </section>
  </MainPage>
)
