export const description = `
# Aquareum: The Video Layer for Everything
## Background

Social networking is in the middle of a decentralized revolution. The world is moving its social structures away from centralized platforms run by megacorporations to federated structures that better align with our values and needs as a society. Projects like Farcaster, Lens, Nostr, and Bluesky's AT Protocol leverage blockchains and public key infrastructure to put control of data firmly in the hands of creators. They all obey the fundamental principle of decentralization: user actions are self-sovereign and immutable.

As these platforms continue to evolve, they will all want rich video functionality. (Some social networks, like TikTok, are themselves 90% rich video functionality.) Video, especially live video, is notoriously difficult and expensive to make work at scale. Legacy platforms such as Twitch, YouTube, and Periscope are massively unprofitable; they only work well due to immense expenditure of servers and bandwidth. Decentralized protocol designers are focused on their protocols, not technical details of video muxing, transcoding, and global low-latency distribution. 

Livepeer and MistServer are great at video muxing, transcoding, and global low-latency video distribution. They have a wide range of rich video features, including Livestreaming, VoD, Clipping, Multistreaming, Transcoding, and AI video generation. All of that content can be efficiently muxed to out millions of users, using low-latency WebRTC. But Livepeer nodes don't have a strong concept of a “user” — who exactly is _allowed_ to stream on a Livepeer node? Livepeer isn't a social network. It's not our job to invent decentralized social.

There's a natural opportunity here — by indexing decentralized social networks and treating those as the source of truth for the world's freely-available video content, we immediately and permissionlessly enable rich video functionality for all of these platforms.

**The end result:**

* **An iOS/Android/Web app that provides a Twitch/YouTube interface, instantly usable by millions of users.**
* **A single-binary node that anyone can run, anywhere.**
* **A set of user-sovereign primitives for expressing creator consent and provenance.**
* **A trustless protocol, such that any app can connect to any node and be confidently served content associated with the dSocial user.**
* **And millions of users that can start watching and creating video content immediately.**

The Aquareum team is seeking funding from the Livepeer Treasury to deliver on this vision. Toward this end, we're seeking **20,000 LPT**. This funding will go toward compensating our team and renting server infrastructure capable of prototyping global video distribution.

## The Team

**Eli Mallon:** Eli is a longtime core contributor to the Livepeer protocol and was the primary architect behind Livepeer Studio. He will be leading the backend, video development, and protocol design.

**Adam Soffer:** Adam was one of the earliest contributors to Livepeer, instrumental in building the Livepeer Subgraph, the Livepeer Explorer, and the Livepeer Studio player and frontend. Adam also created The Web3 Index.  He will be leading the design, frontend development, and indexing layers.

Eli and Adam will be leaving our positions at Livepeer Inc. to take on a vision that we see as core to delivering on the promise of decentralized video. We're excited about participating in the decentralization of Livepeer by establishing another entity in the ecosystem.

## The Plan

### **Aquareum Primitives**

This is the schema. The top-level set of primitives are very familiar: ${"`Segment`"}, ${"`Livestream`"}, ${"`Clip`"}, and ${"`Transcode`"}. Most developers working with these primitives will only need to use a few to compose rich video experiences. They will all be tightly-coupled with React components, making frontend video development fun and easy.

The primitives will be signed by the content creators and include an expressive language for the allowed distribution of their content. Creators will, for example, be able to express licensing information or choose to make content "expire" after a certain amount of time. Aquareum nodes will respect these fields, giving creators control over the lifecycle of their content in the system.
 
**These primitives will be natively anchored into the signing schemes utilized by decentralized social networks. They'll work with Bluesky/Farcaster/etc natively.** Behind the scenes, this set of primitives is expressible as an AT Protocol DAG-CBOR lexicon (for Bluesky) or an EIP-712 schema as utilized by Farcaster and Lens. Additionally, these primitives will be C2PA compliant — they'll respect the industry standard for representing signed video with a coherent provenance chain. This means that C2PA tooling built into video editors and players will be able to confidently inform an end user of the veracity of the video. The EIP-712 schema also makes these primitives suitable for minting onto NFTs. 

The end result is an API that provides you verifiable video content, provided by whichever node that is willing to serve it to you. 

### **Aquareum Indexer**

This app consumes the firehoses from the social networks, and is conceptually similar to a Graph Node subgraph. It then builds all of users' video content into a big distributed database. it  builds an internal state of video clips and livestreams, exposing a GraphQL interface to be consumed by the application. All data served by the indexer is signed and verifiable within its own cryptographic realm.

Our initial approach will be to deliver indexers for Farcaster and AT Protocol, working closely with those communities.

### **Aquareum App**

An app that runs on iPhone, Android, and Web. A familiar frontend, like Twitch or YouTube. Connects to one or more Aquareum Nodes to display a library of video content. All content is accompanied with C2PA metadata — you can always be confident in what you're looking at, even if you don't trust the node you're connected to. All of the video content is associated with, say, ${"`adamsoffer.eth`"} or ${"`@iame.li`"}.

Down the line there will also be creation tools — an embedded livestreaming interface and TikTok-style video editor. While new content will often be created and processed in tandem with an Aquareum node to offload the heavy lifting of video processing, the app will also be fully transcode-enabled, utilizing deterministic WASM-compiled encoding libraries.

### **Aquareum Node**

This bundles the indexer, the app, and the full Livepeer + MistServer software stack into a single executable. It's dead simple to boot it up, configure, and serve content to users over low-latency WebRTC.

It will also facilitate creation of new video content. Users may livestream from the app into Aquareum nodes, and both live and recorded content may be clipped, spliced, composited, and synthesized into new video content. The node leverages the Livepeer Network's GPU processing to do the heavy lifting of video content creation, distribution, and synthesis. However, data is local to the user's app/node until published onto an upstream social network, facilitating both statelessness and moderation. Aquareum nodes will utilize the moderation capabilities built into decentralized social networks to facilitate trust and safety.

## The First Milestone

Aquareum is an ambitious vision, so it's important to be clear about what we're looking to accomplish in this phase.

* First and most boring, we'll be establishing a business entity: Aquareum Inc. Aquareum will take a long time to make, and we're in it for the long haul; this is a necessary step to ensure stability for our team.
* We will deliver the creation of the Aquareum node. This includes a build process for a single statically-linked binary. (This will deliver on a piece of the Livepeer Catalyst vision: a unified version of MistServer and go-livepeer).
* We will deliver a full-stack development environment for Aquareum that's stupid easy to use. Unlike the old Livepeer Catalyst stack, which was opaque and difficult, this will be built from the ground up to encourage community contributions. There will be **one git repository with a single ${"`make`"} command capable of building all components of the project**, and we will be seeking community contributions from Day 1.
* The Aquareum node will run on Windows, macOS, or Linux on AMD64 or ARM64 processors. The Aquareum app will run on iOS, Android, and the Web. More platforms is better, but that's a good start.
* Ship the design of the API primitives, with a focus on signed canonical representations, probably including DAG-CBOR, EIP-712, and C2PA schema.
* Build integrations with our first two social networks: Farcaster and Bluesky.
* 100% of what we build will be open-sourced under an MIT or similarly permissive license.

## Livepeer Network Improvements

So why should the Livepeer Network participants care about this? Aside from the possibility of Aquareum's success driving traffic to the Livepeer network processing, there are numerous specific protocol upgrades:

* We'll be packaging up go-livepeer as a library that can be used within other code, dramatically expanding its use cases. This will make it possible to embed Livepeer Network transcoding in any application capable of linking a C library.
* We will be establishing a GitLab instance with hosted CI runners that will be available free of charge to other projects in the ecosystem. This server will also be used to facilitate project management.
* Aquareum will (finally!) give a coherent answer to how the Livepeer protocol handles scaled video distribution — video users run Aquareum nodes.
* We will seek to upgrade the Livepeer wire protocol to utilize C2PA signatures for transcoding. (This is how it would have been doing originally, if the C2PA had existed when the original protocol was developed.) If necessary, we will deliver the actual protocol improvements via an LIP. This will also dramatically increase the utility of Livepeer transcoding. Transcoded segments are presently mostly opaque. With these schema, transcoded Livepeer segments will instead be digital artifacts with a coherent provenance chain, identifying the user, the transcoder, the content, the licensing, and a million other things.

## Roadmap

We have made our roadmap available on the forums; an archived version is also available at [/ipfs/QmPxASKEjVqWnUpVpmbF35Z9WmAiGKT7sQ6ER5HMuUB9Wt](https://ipfs.io/ipfs/QmPxASKEjVqWnUpVpmbF35Z9WmAiGKT7sQ6ER5HMuUB9Wt).

And one more thing: development will be livestreamed as much as possible — development, code review, project planning will be livestreamed on the Aquareum app and website. We'll be scratching our own itch for a streaming platform while building everything out. We see this as one of the most honest ways to hold ourselves accountable; tune in and see exactly how the treasury funding is being spent! Our livestreaming premiere will be on Tuesday Jul 16 2024 10:00 PDT (12:00 EDT, 17:00 UTC); check the countdown at [aquareum.tv](https://aquareum.tv).

`.trim();
