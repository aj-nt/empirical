package mundane

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aj-nt/empirical/internal/dignity"
)

// ═══════════════════════════════════════════════════════════════════════
// Mundane Interpretation Layer
// ═══════════════════════════════════════════════════════════════════════
//
// Collective-voice interpretations for national charts, ingress charts,
// and transit-to-nation analysis. These are LLM-generated, grounded in
// mundane astrology principles but not expert-authored. Compare against
// traditional sources before relying on them.

// ── Mundane House Meanings ────────────────────────────────────────────

var mundaneHouseMeanings = map[int]string{
	1:  "the nation itself — its people, character, national identity, and general condition. The health and morale of the body politic.",
	2:  "national wealth, the treasury, banks, trade, GDP, and the material resources of the country. National self-worth and economic sovereignty.",
	3:  "communications, media, transportation networks, neighboring countries, trade routes, and the national discourse. The intellectual life of the nation.",
	4:  "the land itself, agriculture, housing, the opposition party, natural resources, mines, and the national foundation. Weather and seismic activity.",
	5:  "national creativity, the arts, entertainment, sports, speculation, the birth rate, children, and national morale. The nation's capacity for joy.",
	6:  "public health, the labor force, civil service, military enlisted ranks, public utilities, and national service. Working conditions and productivity.",
	7:  "foreign relations, treaties, alliances, open enemies, trade partners, and marriage/divorce rates. The nation's relationship with other nations.",
	8:  "national debt, taxation, shared resources with other nations, death rates, inheritance, and transformative crises. The nation's relationship with power and mortality.",
	9:  "higher education, the courts, philosophy, religion, foreign travel, publishing, and the nation's worldview. Long-distance trade and international law.",
	10: "the government, the executive branch, the ruling party, national reputation, and authority figures. The nation's standing in the world.",
	11: "the legislature, parliament, congress, political parties, social movements, the national budget, and collective aspirations. The people as a political body.",
	12: "hospitals, prisons, asylums, secret services, hidden enemies, espionage, and national secrets. The collective unconscious and national shadow.",
}

// ── Mundane Planet Meanings ───────────────────────────────────────────

var mundanePlanetMeanings = map[string]string{
	"Sun":     "the executive, the head of state, national identity, and the ruling authority. The nation's vitality and sense of purpose",
	"Moon":    "the general public, popular sentiment, the mood of the nation, women, and domestic affairs. The collective emotional state",
	"Mercury": "communications, media, trade, transportation, diplomacy, and the intellectual climate. The national conversation",
	"Venus":   "alliances, diplomacy, trade agreements, cultural influence, national values, and soft power. The arts and social harmony",
	"Mars":    "the military, police, national aggression, labor disputes, industrial action, and collective anger. The capacity for force",
	"Jupiter": "the judiciary, religious institutions, higher education, foreign expansion, national optimism, and economic growth. The nation's faith and reach",
	"Saturn":  "government structure, law enforcement, national discipline, economic contraction, austerity, and institutional authority. The nation's limits and responsibilities",
	"Uranus":  "revolution, technological change, social upheaval, innovation, sudden political shifts, and generational breaks. The nation's capacity for disruption",
	"Neptune": "national ideals, propaganda, collective delusion, the arts, spiritual movements, and scandals. The nation's dreams and deceptions",
	"Pluto":   "national power, transformation, intelligence agencies, organized crime, debt, and existential crises. The nation's relationship with power and death",
	"Node":    "the nation's evolutionary direction, collective destiny, and the point of amplification. Where the country is being pulled",
	"SouthNode": "the nation's default patterns, inherited structures, and the point of release. What the country already knows and must move beyond",
}

// ── Mundane Planet-in-House Interpretations ────────────────────────────

var mundanePlanetHouseMeanings = map[string]map[int]string{
	"Sun": {
		1:  "Sun in the 1st house: the executive and national identity are fused. The head of state embodies the nation's character. Strong national self-image and a need for international recognition. The country projects confidence and demands to be seen. The shadow is national narcissism — the belief that the nation's interests are the world's interests.",
		2:  "Sun in the 2nd house: national identity invested in economic strength. The country defines itself through wealth, trade, and material power. Economic indicators become measures of national worth. The shadow is reducing national purpose to GDP and treating economic competitors as existential threats.",
		3:  "Sun in the 3rd house: national identity expressed through communication and information. The country sees itself as a voice — in media, technology, and discourse. Neighboring countries and trade routes shape national self-image. The shadow is information warfare and the belief that controlling the narrative is controlling reality.",
		4:  "Sun in the 4th house: national identity rooted in the land and the people's private lives. The country defines itself through territory, agriculture, and domestic stability. The opposition party holds unusual influence over national direction. The shadow is isolationism and the belief that the nation's borders are sacred and inviolable.",
		5:  "Sun in the 5th house: national identity expressed through creativity, spectacle, and national pride. The country sees itself as a cultural force. Sports, entertainment, and national celebrations become expressions of identity. The shadow is nationalism as performance and the need for constant external validation.",
		6:  "Sun in the 6th house: national identity expressed through work, service, and public health. The country defines itself through its labor force and its capacity to solve problems. Civil service and military enlisted ranks carry unusual weight. The shadow is reducing national purpose to productivity and treating citizens as resources.",
		7:  "Sun in the 7th house: national identity discovered through relationship with other nations. The country defines itself through alliances, treaties, and its role in the international order. Foreign policy is not just policy — it is identity. The shadow is losing national sovereignty to alliance obligations.",
		8:  "Sun in the 8th house: national identity forged through crisis, debt, and transformation. The country's sense of self is shaped by what it survives. National debt, intelligence operations, and existential threats define the national character. The shadow is a nation that only knows itself through conflict.",
		9:  "Sun in the 9th house: national identity as a mission. The country sees itself as having a purpose beyond its borders — ideological, religious, or civilizational. Courts, universities, and belief systems shape national direction. The shadow is crusading foreign policy and the belief that the nation has a divine mandate.",
		10: "Sun in the 10th house: national identity as global standing. The country defines itself through its reputation, its government's authority, and its position in the international hierarchy. The executive branch dominates national life. The shadow is authoritarianism — the belief that the government IS the nation.",
		11: "Sun in the 11th house: national identity expressed through collective movements and democratic institutions. The country sees itself as a project of the people. Congress, social movements, and the national budget are the theaters of identity. The shadow is populism — the belief that the loudest voice speaks for all.",
		12: "Sun in the 12th house: national identity that is hidden, complex, and not fully visible to itself. The country's true character operates below the surface — in intelligence agencies, secret diplomacy, and the national unconscious. The shadow is a nation that doesn't know what it is and acts from hidden compulsions.",
	},
	"Moon": {
		1:  "Moon in the 1st house: the public mood IS the national condition. Popular sentiment directly shapes national direction. The country is emotionally reactive — its mood swings are visible to the world. The people feel personally invested in the nation's image. The shadow is mob rule and policy driven by feeling rather than strategy.",
		2:  "Moon in the 2nd house: public sentiment tied to economic conditions. The national mood rises and falls with markets, employment, and material security. Consumer confidence is a leading indicator of national stability. The shadow is economic anxiety driving political decisions and the public measuring national worth by their personal finances.",
		3:  "Moon in the 3rd house: public sentiment expressed through media and communication. The national mood is shaped by what is said and how it is said. Social media, news cycles, and public discourse are the pulse of the nation. The shadow is a population that feels everything and processes nothing.",
		4:  "Moon in the 4th house: public sentiment rooted in the land, housing, and domestic life. The national mood is shaped by home, family, and the sense of belonging. Agricultural conditions and housing markets drive public feeling. The shadow is nostalgia politics — the belief that the past was better and must be restored.",
		5:  "Moon in the 5th house: public sentiment expressed through culture, entertainment, and national pride. The national mood is shaped by sports victories, cultural exports, and collective celebrations. The shadow is bread and circuses — keeping the public distracted while decisions are made elsewhere.",
		6:  "Moon in the 6th house: public sentiment tied to work, health, and service. The national mood is shaped by employment conditions, healthcare, and the experience of daily labor. Labor movements and public health crises drive public feeling. The shadow is a population that defines itself by its suffering.",
		7:  "Moon in the 7th house: public sentiment shaped by foreign relations. The national mood is calibrated against other nations — alliances, enemies, and international standing. The people feel personally about foreign policy. The shadow is xenophobia and the belief that national problems are caused by outsiders.",
		8:  "Moon in the 8th house: public sentiment that is intense, fearful, and focused on existential threats. The national mood is shaped by debt, crisis, and what is hidden. The people are anxious about what they cannot see. The shadow is collective paranoia and the belief that enemies are everywhere.",
		9:  "Moon in the 9th house: public sentiment shaped by ideology, religion, and the national worldview. The national mood is tied to belief systems and the sense of national purpose. Courts and universities shape public feeling. The shadow is ideological purity tests and the belief that disagreement is heresy.",
		10: "Moon in the 10th house: public sentiment focused on the government and national leadership. The national mood is shaped by approval ratings, executive actions, and the visible exercise of authority. The people watch the government closely. The shadow is a population that defines itself by who it hates in power.",
		11: "Moon in the 11th house: public sentiment expressed through collective action and social movements. The national mood is shaped by Congress, protests, and the sense of collective possibility. The people feel like participants in history. The shadow is revolutionary fervor that burns out before it builds anything.",
		12: "Moon in the 12th house: public sentiment that is hidden, confused, and operating below awareness. The national mood is shaped by what is not said — secrets, scandals, and the collective unconscious. The people feel something is wrong but cannot name it. The shadow is a nation that acts from unexamined fear.",
	},
	"Mercury": {
		1:  "Mercury in the 1st house: the national conversation IS the national identity. The country is perceived through its media, its rhetoric, and its information environment. Communication shapes how the nation sees itself and is seen. The shadow is a nation that talks constantly but says nothing.",
		2:  "Mercury in the 2nd house: the national conversation focused on economics. Trade negotiations, market analysis, and financial communication dominate. The country thinks in terms of value and exchange. The shadow is reducing all discourse to economic terms and treating information as a commodity.",
		3:  "Mercury in the 3rd house: the national conversation at its most natural. Media, transportation, and neighboring relations are the primary theaters of national thought. The country is a hub of information and movement. The shadow is information overload — so much communication that nothing lands.",
		4:  "Mercury in the 4th house: the national conversation focused on domestic affairs. Land use, housing policy, and the opposition's messaging dominate. The country thinks about its foundation. The shadow is provincial thinking — the belief that local concerns are universal concerns.",
		5:  "Mercury in the 5th house: the national conversation expressed through culture and entertainment. The country thinks through art, sports, and spectacle. Creative industries shape national discourse. The shadow is celebrity culture replacing political discourse.",
		6:  "Mercury in the 6th house: the national conversation focused on work, health, and administration. The country thinks about systems, efficiency, and public health. Civil service communications dominate. The shadow is bureaucratic language that obscures rather than clarifies.",
		7:  "Mercury in the 7th house: the national conversation shaped by foreign relations. Diplomacy, treaty negotiations, and international media shape national thought. The country thinks through its relationships. The shadow is a nation that cannot form an opinion without checking with its allies.",
		8:  "Mercury in the 8th house: the national conversation focused on what is hidden. Intelligence, debt, and investigative journalism dominate. The country thinks about secrets and power. The shadow is conspiracy thinking — the belief that the real story is always hidden.",
		9:  "Mercury in the 9th house: the national conversation focused on ideology and law. Court decisions, academic discourse, and religious debate shape national thought. The country thinks in principles. The shadow is legalism and the belief that the right argument wins regardless of consequences.",
		10: "Mercury in the 10th house: the national conversation focused on government and leadership. Executive communications, policy announcements, and political rhetoric dominate. The country thinks through its leaders. The shadow is propaganda — communication that serves power rather than truth.",
		11: "Mercury in the 11th house: the national conversation expressed through collective deliberation. Congressional debate, social movements, and public forums shape national thought. The country thinks together. The shadow is groupthink and the belief that consensus is truth.",
		12: "Mercury in the 12th house: the national conversation that is hidden, coded, or operating below the surface. Intelligence communications, back-channel diplomacy, and the unspoken shape national thought. The country thinks in secrets. The shadow is a nation that cannot speak honestly about itself.",
	},
	"Venus": {
		1:  "Venus in the 1st house: national values and alliances worn on the surface. The country is perceived as diplomatic, cultured, and attractive to partners. Soft power is the primary mode of national expression. The shadow is a nation that prioritizes being liked over being respected.",
		2:  "Venus in the 2nd house: national values expressed through economic relationships. Trade agreements, cultural exports, and financial diplomacy define the country's approach to value. The nation attracts wealth through relationship. The shadow is treating allies as customers and values as negotiable.",
		3:  "Venus in the 3rd house: national values expressed through communication and media. The country's cultural output — film, music, literature — shapes its international standing. Diplomatic language is an art form. The shadow is soft power that substitutes for substance.",
		4:  "Venus in the 4th house: national values rooted in domestic life and the land. The country's relationship with its territory, its agricultural beauty, and its domestic culture define its values. The shadow is valuing the land over the people who live on it.",
		5:  "Venus in the 5th house: national values expressed through culture, arts, and celebration. The country is a cultural exporter — its entertainment, fashion, and sports define its global image. The shadow is cultural imperialism and the belief that the nation's taste is universal taste.",
		6:  "Venus in the 6th house: national values expressed through labor relations and public health. The country's treatment of workers, its healthcare system, and its civil service define its values. The shadow is performative compassion — looking caring without being caring.",
		7:  "Venus in the 7th house: national values expressed through alliances and partnerships. The country defines itself through its relationships — treaties, trade partners, and diplomatic ties. The shadow is alliance dependency and the inability to act without consensus.",
		8:  "Venus in the 8th house: national values forged through financial entanglement and shared resources. The country's relationship with debt, international finance, and economic power defines its values. The shadow is financial coercion disguised as partnership.",
		9:  "Venus in the 9th house: national values expressed through ideology and international law. The country's legal system, academic culture, and religious tolerance define its global standing. The shadow is moral imperialism — the belief that the nation's values are universal values.",
		10: "Venus in the 10th house: national values expressed through government and international reputation. The country's leadership is judged by its diplomacy, its cultural sophistication, and its ability to attract allies. The shadow is a government that prioritizes image over governance.",
		11: "Venus in the 11th house: national values expressed through collective institutions and social movements. Congress, civil society, and the national budget reflect what the country truly values. The shadow is values as performance — saying the right things while funding the wrong things.",
		12: "Venus in the 12th house: national values that are hidden, compromised, or operating in secret. The country's true values may differ from its stated values. Secret alliances, hidden financial relationships, and unacknowledged cultural influences shape national direction. The shadow is a nation that doesn't know what it stands for.",
	},
	"Mars": {
		1:  "Mars in the 1st house: military power and national aggression worn on the surface. The country is perceived as forceful, competitive, and willing to fight. The military shapes national identity. The shadow is militarism — the belief that force is the first and best answer.",
		2:  "Mars in the 2nd house: national aggression directed at economic targets. Trade wars, economic sanctions, and resource competition define the country's approach to conflict. The military budget is a statement of values. The shadow is economic warfare that destroys what it claims to protect.",
		3:  "Mars in the 3rd house: national aggression expressed through communication and information warfare. The country fights with words, propaganda, and cyber operations. Media is a battlefield. The shadow is a nation that cannot distinguish between argument and attack.",
		4:  "Mars in the 4th house: national aggression rooted in territorial defense and domestic conflict. The country is prepared to fight for its land, its borders, and its internal order. The opposition party may be militant. The shadow is civil conflict and the militarization of domestic politics.",
		5:  "Mars in the 5th house: national aggression expressed through competition, sports, and cultural battles. The country fights on fields, in stadiums, and in culture wars. National pride is expressed through victory. The shadow is a nation that treats everything as a contest to be won.",
		6:  "Mars in the 6th house: national aggression directed at internal enemies and labor conflicts. The country fights through strikes, police actions, and public health enforcement. The military's enlisted ranks and civil service are theaters of conflict. The shadow is a nation at war with its own workforce.",
		7:  "Mars in the 7th house: national aggression directed at foreign enemies and alliance conflicts. The country defines itself through its adversaries. Military alliances are both shield and sword. The shadow is a nation that needs an enemy to know who it is.",
		8:  "Mars in the 8th house: national aggression directed at existential threats and hidden enemies. The country fights through intelligence operations, covert action, and financial warfare. The shadow is a nation that fights wars it cannot acknowledge.",
		9:  "Mars in the 9th house: national aggression expressed through ideological conflict and legal warfare. The country fights for beliefs, through courts, and across borders. The shadow is crusading — the belief that military force serves a higher purpose.",
		10: "Mars in the 10th house: national aggression expressed through the executive and military command. The country's leadership is defined by its willingness to use force. The commander-in-chief role dominates. The shadow is a government that governs through threat.",
		11: "Mars in the 11th house: national aggression expressed through collective action and political conflict. The country fights through Congress, protests, and social movements. The shadow is revolutionary violence and the belief that the ends justify the means.",
		12: "Mars in the 12th house: national aggression that is hidden, covert, or turned inward. The country fights through intelligence agencies, secret operations, and psychological warfare. The shadow is a nation that is its own worst enemy — self-sabotage through unacknowledged aggression.",
	},
	"Jupiter": {
		1:  "Jupiter in the 1st house: national expansion and optimism worn on the surface. The country is perceived as generous, confident, and growing. The nation's reach exceeds its grasp — and that is the point. The shadow is imperial overreach and the belief that bigger is always better.",
		2:  "Jupiter in the 2nd house: national expansion through economic growth. The country attracts wealth, investment, and material abundance. Trade expansion and financial optimism define the national mood. The shadow is speculative bubbles and the belief that growth is permanent.",
		3:  "Jupiter in the 3rd house: national expansion through communication, education, and transportation. The country's media, universities, and infrastructure grow. The national conversation is optimistic and far-reaching. The shadow is information sprawl — more communication, less understanding.",
		4:  "Jupiter in the 4th house: national expansion through land, housing, and domestic growth. The country grows from within — population, territory, and domestic institutions expand. The shadow is suburban sprawl and the belief that growth must consume land.",
		5:  "Jupiter in the 5th house: national expansion through culture, creativity, and national pride. The country's arts, entertainment, and sports achieve global reach. The birth rate and national morale are high. The shadow is cultural excess and the belief that more spectacle means more substance.",
		6:  "Jupiter in the 6th house: national expansion through labor, health systems, and public administration. The country grows its workforce, its healthcare capacity, and its civil service. The shadow is bureaucratic bloat and the belief that more government means better government.",
		7:  "Jupiter in the 7th house: national expansion through alliances and international partnerships. The country grows its network of allies, trade partners, and diplomatic relationships. The shadow is alliance overextension and the belief that more friends means more security.",
		8:  "Jupiter in the 8th house: national expansion through debt, shared resources, and financial power. The country grows through leverage — borrowing, investing, and controlling financial flows. The shadow is debt-fueled growth and the belief that tomorrow's problems can be paid for with tomorrow's borrowing.",
		9:  "Jupiter in the 9th house: national expansion through ideology, law, and global influence. The country's worldview, legal system, and cultural values achieve international reach. The shadow is ideological imperialism and the belief that the nation's way is the only way.",
		10: "Jupiter in the 10th house: national expansion through government and international standing. The country's leadership achieves global prominence. The executive branch grows in power and reach. The shadow is executive overreach and the belief that strong leadership means unchecked leadership.",
		11: "Jupiter in the 11th house: national expansion through democratic institutions and collective movements. Congress, social movements, and the national budget grow in scope and ambition. The shadow is legislative bloat and the belief that more laws mean more justice.",
		12: "Jupiter in the 12th house: national expansion through hidden channels — intelligence, spirituality, and the national unconscious. The country grows in ways that are not immediately visible. The shadow is expansion of the surveillance state and the belief that security requires secrecy.",
	},
	"Saturn": {
		1:  "Saturn in the 1st house: national discipline and limitation worn on the surface. The country is perceived as serious, structured, and enduring. National identity is forged through hardship and responsibility. The shadow is national pessimism and the belief that the country's best days are behind it.",
		2:  "Saturn in the 2nd house: national discipline expressed through economic constraint. The country faces limits on wealth, trade, and material resources. Austerity, budget discipline, and economic restructuring define the period. The shadow is economic contraction as identity and the belief that scarcity is permanent.",
		3:  "Saturn in the 3rd house: national discipline expressed through communication and education. The country faces limits on media, information flow, and intellectual freedom. Censorship, educational reform, and communication infrastructure are tested. The shadow is a nation that controls what can be said.",
		4:  "Saturn in the 4th house: national discipline expressed through land, housing, and domestic structure. The country faces limits on territory, housing availability, and domestic stability. The opposition party gains structural power. The shadow is a nation that walls itself in.",
		5:  "Saturn in the 5th house: national discipline expressed through cultural constraint and demographic decline. The country faces limits on creative expression, entertainment, and population growth. The birth rate falls. The shadow is cultural austerity and the belief that joy is frivolous.",
		6:  "Saturn in the 6th house: national discipline expressed through labor, health, and public administration. The country faces limits on workforce capacity, healthcare systems, and civil service. Labor disputes and public health crises test national resilience. The shadow is a nation that works itself to exhaustion.",
		7:  "Saturn in the 7th house: national discipline expressed through alliances and foreign relations. The country faces limits on its international partnerships. Treaties are tested, alliances strained, and diplomatic relationships require work. The shadow is isolation and the belief that the nation must stand alone.",
		8:  "Saturn in the 8th house: national discipline expressed through debt, taxation, and existential limits. The country faces hard truths about what it owes and what it controls. Financial restructuring and power transitions are forced. The shadow is a nation that cannot escape its debts.",
		9:  "Saturn in the 9th house: national discipline expressed through law, ideology, and the national worldview. The country faces limits on its belief systems. Courts are tested, universities face scrutiny, and national ideology confronts reality. The shadow is ideological rigidity and the belief that questioning the faith is treason.",
		10: "Saturn in the 10th house: national discipline expressed through government and leadership. The country faces limits on executive power and national reputation. The government is tested, held accountable, and forced to deliver. The shadow is authoritarian response to limitation — tightening control when challenged.",
		11: "Saturn in the 11th house: national discipline expressed through democratic institutions and collective responsibility. Congress, social movements, and the national budget face limits. The people are asked to sacrifice. The shadow is democratic erosion and the belief that strong leadership is better than messy democracy.",
		12: "Saturn in the 12th house: national discipline expressed through hidden structures and institutional secrets. The country faces limits on what it can hide. Intelligence agencies, prisons, and secret programs are exposed or restructured. The shadow is a nation that builds walls around its own darkness.",
	},
	"Uranus": {
		1:  "Uranus in the 1st house: national disruption worn on the surface. The country is perceived as unpredictable, innovative, and breaking from its own past. National identity undergoes sudden change. The shadow is instability as identity and the belief that all tradition must be destroyed.",
		2:  "Uranus in the 2nd house: national disruption of the economy. Sudden changes in markets, trade relationships, and national wealth. New economic models emerge. Cryptocurrency, alternative finance, and economic revolution. The shadow is economic chaos and the belief that disruption is always progress.",
		3:  "Uranus in the 3rd house: national disruption of communication and information. New media platforms, sudden changes in public discourse, and technological revolution in how the country talks to itself. The shadow is information warfare and the belief that breaking the old media is the same as building something better.",
		4:  "Uranus in the 4th house: national disruption of land, housing, and domestic order. Sudden changes in territory, property rights, and the relationship between people and place. The opposition party may suddenly gain or lose power. The shadow is displacement and the belief that the old foundations were never valid.",
		5:  "Uranus in the 5th house: national disruption of culture, creativity, and national morale. Sudden changes in entertainment, arts, and what the nation celebrates. New cultural forms emerge violently. The shadow is cultural chaos and the belief that all tradition is oppression.",
		6:  "Uranus in the 6th house: national disruption of labor, health, and public services. Sudden changes in work patterns, healthcare delivery, and civil service structure. Automation, labor revolts, and public health innovations. The shadow is worker displacement and the belief that efficiency justifies human cost.",
		7:  "Uranus in the 7th house: national disruption of alliances and foreign relations. Sudden changes in treaties, partnerships, and the international order. Old allies become adversaries; old enemies become partners. The shadow is diplomatic chaos and the belief that all commitments are provisional.",
		8:  "Uranus in the 8th house: national disruption of debt, power structures, and hidden systems. Sudden changes in financial systems, intelligence operations, and the nation's relationship with power. The shadow is financial collapse and the belief that the old power structures must be destroyed before new ones can be built.",
		9:  "Uranus in the 9th house: national disruption of ideology, law, and the national worldview. Sudden changes in legal systems, educational institutions, and national belief. The shadow is ideological chaos and the belief that all received wisdom is oppression.",
		10: "Uranus in the 10th house: national disruption of government and leadership. Sudden changes in executive power, national reputation, and who holds authority. Coups, resignations, and unexpected leaders. The shadow is governmental collapse and the belief that tearing down is the same as building.",
		11: "Uranus in the 11th house: national disruption of democratic institutions and collective movements. Sudden changes in Congress, social movements, and the national budget. Revolutionary political realignments. The shadow is mob rule and the belief that the loudest revolution is the most legitimate.",
		12: "Uranus in the 12th house: national disruption of hidden systems and the national unconscious. Sudden exposure of secrets, intelligence failures, and institutional breakdown. The shadow is a nation that doesn't know what it's becoming and acts from unexamined impulse.",
	},
	"Neptune": {
		1:  "Neptune in the 1st house: national identity that is fluid, idealistic, and hard to define. The country is perceived through its myths, its dreams, and its propaganda. National image is a projection screen for others' hopes and fears. The shadow is a nation that doesn't know what it is and believes its own lies.",
		2:  "Neptune in the 2nd house: national values and economy that are fluid, speculative, and potentially illusory. The country's relationship with wealth is shaped by bubbles, fraud, and financial mystification. The shadow is economic delusion and the belief that printing money creates value.",
		3:  "Neptune in the 3rd house: national communication that is poetic, deceptive, and boundaryless. The country's media environment is shaped by propaganda, misinformation, and inspired rhetoric. The shadow is a nation that cannot distinguish between truth and narrative.",
		4:  "Neptune in the 4th house: national relationship with land and home that is idealized, nostalgic, and potentially delusional. The country's sense of territory is shaped by myth rather than reality. The shadow is a nation that fights for a homeland that never existed.",
		5:  "Neptune in the 5th house: national culture that is transcendent, glamorous, and potentially escapist. The country's arts, entertainment, and national celebrations achieve mythic status. The shadow is a nation that prefers its fantasy to its reality.",
		6:  "Neptune in the 6th house: national labor and health systems that are compassionate, confused, and potentially dysfunctional. The country's relationship with work and health is shaped by idealism and neglect. The shadow is a nation that means well but cannot deliver.",
		7:  "Neptune in the 7th house: national alliances that are idealized, deceptive, and potentially delusional. The country's relationships with other nations are shaped by projection and wishful thinking. The shadow is a nation that trusts the wrong allies and distrusts the right ones.",
		8:  "Neptune in the 8th house: national relationship with debt and power that is obscured, mystical, and potentially fraudulent. The country's financial system and intelligence operations operate in fog. The shadow is a nation whose true power structure is invisible even to itself.",
		9:  "Neptune in the 9th house: national ideology that is spiritual, universal, and potentially ungrounded. The country's belief systems, legal framework, and worldview are shaped by idealism and confusion. The shadow is a nation that believes its own mythology and acts on faith without evidence.",
		10: "Neptune in the 10th house: national leadership that is charismatic, elusive, and potentially deceptive. The country's government and reputation are shaped by image, scandal, and the gap between appearance and reality. The shadow is a nation led by people who believe their own propaganda.",
		11: "Neptune in the 11th house: national collective movements that are utopian, confused, and potentially self-defeating. The country's democratic institutions and social movements are shaped by idealism and disillusionment. The shadow is a nation that dreams of revolution but cannot organize a government.",
		12: "Neptune in the 12th house: national unconscious that is oceanic, creative, and potentially overwhelming. The country's hidden systems, secrets, and collective psyche are the primary theater of national life. The shadow is a nation that is more real in its imagination than in its actions.",
	},
	"Pluto": {
		1:  "Pluto in the 1st house: national power and transformation worn on the surface. The country is perceived as intense, magnetic, and not fully knowable. National identity is forged through crisis and regeneration. The shadow is a nation that only knows itself through destruction and rebirth.",
		2:  "Pluto in the 2nd house: national power expressed through economic control and resource transformation. The country's relationship with wealth is all-or-nothing — cycles of accumulation and loss. The shadow is economic coercion and the belief that controlling resources is controlling destiny.",
		3:  "Pluto in the 3rd house: national power expressed through information control and communication warfare. The country's media, intelligence, and information environment are theaters of power. The shadow is a nation that controls what can be known and punishes those who know too much.",
		4:  "Pluto in the 4th house: national power rooted in land, territory, and the domestic underworld. The country's relationship with its own ground is intense, contested, and subject to transformation. The shadow is a nation that will destroy its own land to control it.",
		5:  "Pluto in the 5th house: national power expressed through culture, spectacle, and demographic transformation. The country's arts, entertainment, and national identity undergo cycles of death and rebirth. The shadow is a nation that consumes its own culture and calls it renewal.",
		6:  "Pluto in the 6th house: national power expressed through labor control, health systems, and institutional transformation. The country's workforce, public health, and civil service are sites of power struggle. The shadow is a nation that treats its people as resources to be extracted.",
		7:  "Pluto in the 7th house: national power expressed through alliance domination and adversarial relationships. The country's foreign relations are intense, controlling, and subject to cycles of merger and rupture. The shadow is a nation that consumes its allies and annihilates its enemies.",
		8:  "Pluto in the 8th house: national power at its most natural — debt, intelligence, and existential transformation. The country's relationship with power itself is the central national drama. The shadow is a nation that is more comfortable with destruction than with peace.",
		9:  "Pluto in the 9th house: national power expressed through ideological control and legal transformation. The country's belief systems, courts, and worldview are sites of power struggle. The shadow is a nation that enforces belief through law and punishes heresy.",
		10: "Pluto in the 10th house: national power expressed through government and leadership transformation. The country's executive branch, national reputation, and authority structures undergo cycles of collapse and regeneration. The shadow is a nation that cycles through strongmen and cannot sustain democratic transition.",
		11: "Pluto in the 11th house: national power expressed through collective transformation and institutional destruction. The country's Congress, social movements, and democratic institutions are sites of power struggle. The shadow is a nation that destroys its own democracy and calls it renewal.",
		12: "Pluto in the 12th house: national power that is hidden, operating through secret systems and the collective unconscious. The country's true power structure is invisible. Intelligence agencies, organized crime, and hidden networks shape national direction. The shadow is a nation controlled by forces it cannot name.",
	},
	"Node": {
		1:  "North Node in the 1st house: the nation's evolutionary direction is toward a stronger, more visible national identity. The country is being pulled to assert itself, to be seen, to stand alone. Planets conjunct this point get louder. The growth edge is national self-definition without arrogance.",
		2:  "North Node in the 2nd house: the nation's evolutionary direction is toward economic self-definition. The country is being pulled to define its values, build its wealth, and establish material sovereignty. The growth edge is prosperity that serves the people rather than defining them.",
		3:  "North Node in the 3rd house: the nation's evolutionary direction is toward communication, education, and information. The country is being pulled to become a voice — in media, technology, and discourse. The growth edge is communication that informs rather than manipulates.",
		4:  "North Node in the 4th house: the nation's evolutionary direction is toward domestic strength and territorial integrity. The country is being pulled to secure its foundation — land, housing, and the private lives of its people. The growth edge is a nation that is strong at home without being closed to the world.",
		5:  "North Node in the 5th house: the nation's evolutionary direction is toward cultural expression and national confidence. The country is being pulled to create, to celebrate, to express its identity through art and joy. The growth edge is national pride that doesn't require external validation.",
		6:  "North Node in the 6th house: the nation's evolutionary direction is toward service, health, and productive capacity. The country is being pulled to build systems that work — healthcare, labor, civil service. The growth edge is a nation that serves its people without reducing them to their function.",
		7:  "North Node in the 7th house: the nation's evolutionary direction is toward partnership and alliance. The country is being pulled to define itself through relationship — treaties, diplomacy, and international cooperation. The growth edge is partnership that strengthens rather than replaces national sovereignty.",
		8:  "North Node in the 8th house: the nation's evolutionary direction is toward transformation through crisis and shared power. The country is being pulled to confront its debts, its secrets, and its relationship with power itself. The growth edge is transformation that liberates rather than destroys.",
		9:  "North Node in the 9th house: the nation's evolutionary direction is toward ideological leadership and global influence. The country is being pulled to define its beliefs, expand its worldview, and project its values. The growth edge is influence that persuades rather than imposes.",
		10: "North Node in the 10th house: the nation's evolutionary direction is toward governmental authority and international standing. The country is being pulled to lead, to govern, to be recognized. The growth edge is authority earned through service rather than seized through force.",
		11: "North Node in the 11th house: the nation's evolutionary direction is toward collective governance and democratic expression. The country is being pulled to build institutions that represent the people. The growth edge is democracy that includes rather than divides.",
		12: "North Node in the 12th house: the nation's evolutionary direction is toward spiritual depth and institutional transparency. The country is being pulled to examine its hidden self — its secrets, its shadow, its unconscious. The growth edge is a nation that knows itself fully and acts from that knowledge.",
	},
	"SouthNode": {
		1:  "South Node in the 1st house: the nation's default pattern is excessive self-focus. The country already knows how to assert itself, to stand alone, to project a finished identity. The release is learning that national identity doesn't need to be demonstrated to be real.",
		2:  "South Node in the 2nd house: the nation's default pattern is economic accumulation as identity. The country already knows how to build wealth, to measure worth in material terms. The release is learning that national value is not the same as national wealth.",
		3:  "South Node in the 3rd house: the nation's default pattern is communication without substance. The country already knows how to talk, to inform, to fill the airwaves. The release is learning that not everything needs to be said or known.",
		4:  "South Node in the 4th house: the nation's default pattern is territorial fixation. The country already knows how to defend its land, to build walls, to protect its domestic sphere. The release is learning that the nation's foundation is its people, not its borders.",
		5:  "South Node in the 5th house: the nation's default pattern is cultural performance as identity. The country already knows how to entertain, to celebrate, to project confidence. The release is learning that national pride doesn't need an audience.",
		6:  "South Node in the 6th house: the nation's default pattern is productivity as purpose. The country already knows how to work, to serve, to optimize. The release is learning that the nation's worth is not its output.",
		7:  "South Node in the 7th house: the nation's default pattern is alliance dependency. The country already knows how to partner, to accommodate, to merge. The release is learning that the nation is already sovereign and doesn't need a mirror.",
		8:  "South Node in the 8th house: the nation's default pattern is power through crisis. The country already knows how to survive, to control, to keep secrets. The release is learning that the nation's power is not measured by what it destroys.",
		9:  "South Node in the 9th house: the nation's default pattern is ideological certainty. The country already knows what it believes, what it stands for, what it exports. The release is learning that the meaning is already here and doesn't need to be enforced.",
		10: "South Node in the 10th house: the nation's default pattern is governmental dominance. The country already knows how to lead, to command, to project authority. The release is learning that the government serves the nation, not the reverse.",
		11: "South Node in the 11th house: the nation's default pattern is collective identity as substitute for individual thought. The country already knows how to organize, to mobilize, to belong. The release is learning that the collective is made of individuals who must be free.",
		12: "South Node in the 12th house: the nation's default pattern is hidden power and unexamined shadow. The country already knows how to operate in secret, to transcend accountability, to dissolve boundaries. The release is learning that the nation can have edges and still be whole.",
	},
}

// ── Mundane Transit Interpretations ────────────────────────────────────

// InterpretMundaneTransit returns a natural-language interpretation of a
// transit from a transiting planet to a natal planet in a national chart.
func InterpretMundaneTransit(transitPlanet, natalPlanet, aspect string, orb float64) string {
	tp, tok := mundanePlanetMeanings[transitPlanet]
	if !tok {
		tp = strings.ToLower(transitPlanet)
	}
	np, nok := mundanePlanetMeanings[natalPlanet]
	if !nok {
		np = strings.ToLower(natalPlanet)
	}

	aspectMeanings := map[string]string{
		"conjunction": "merge and amplify — the transiting energy fuses with the national placement. A new cycle begins.",
		"opposition":  "polarize and confront — the transiting energy challenges the national placement from across the sky. A crisis of balance.",
		"square":      "friction and forced growth — the transiting energy creates tension with the national placement. Development through conflict.",
		"trine":       "flow and support — the transiting energy harmonizes with the national placement. Opportunity without resistance.",
		"sextile":     "opportunity and invitation — the transiting energy opens a door for the national placement. Action required to receive the benefit.",
	}

	am, ok := aspectMeanings[aspect]
	if !ok {
		am = fmt.Sprintf("forms a %s aspect", aspect)
	}

	return fmt.Sprintf("Transiting %s (%s) in %s to natal %s (%s): %s",
		transitPlanet, tp, aspect, natalPlanet, np, am)
}

// ── Mundane Ingress Chart Interpretations ──────────────────────────────

// InterpretIngressASC returns an interpretation of the Ascendant in an
// ingress chart cast for a specific location.
func InterpretIngressASC(ascSign string, ascDegree float64) string {
	meanings := map[string]string{
		"Aries":       "ASC Aries: the period begins with initiatory, martial energy. The nation acts first and asks questions later. Leadership is direct, competitive, and potentially confrontational. Military and executive action dominate the ingress period. The risk is impulsiveness — acting before the situation is understood.",
		"Taurus":      "ASC Taurus: the period begins with steady, accumulative energy. The nation focuses on economic stability, material security, and what endures. Slow, deliberate action. The risk is stubbornness — refusing to adapt when adaptation is required.",
		"Gemini":      "ASC Gemini: the period begins with communicative, dualistic energy. The nation is in conversation with itself and the world. Media, trade, and information dominate. The risk is scattered attention — too many priorities, not enough follow-through.",
		"Cancer":      "ASC Cancer: the period begins with protective, domestic energy. The nation turns inward — toward home, family, and national security. The public mood drives policy. The risk is emotional reactivity — policy driven by feeling rather than strategy.",
		"Leo":         "ASC Leo: the period begins with radiant, commanding energy. The nation demands attention and projects confidence. The executive and national identity are central. The risk is performative leadership — looking strong rather than being strong.",
		"Virgo":       "ASC Virgo: the period begins with analytical, service-oriented energy. The nation focuses on systems, efficiency, and problem-solving. Civil service and public health dominate. The risk is over-analysis — perfecting the plan while the moment passes.",
		"Libra":       "ASC Libra: the period begins with diplomatic, relational energy. The nation focuses on alliances, treaties, and international standing. Soft power and negotiation dominate. The risk is indecision — seeking consensus when decisive action is needed.",
		"Scorpio":     "ASC Scorpio: the period begins with intense, transformative energy. The nation confronts hidden threats, financial power, and existential questions. Intelligence and debt are central. The risk is paranoia — seeing enemies where there are only challenges.",
		"Sagittarius": "ASC Sagittarius: the period begins with expansive, ideological energy. The nation looks outward — toward foreign policy, legal frameworks, and belief systems. The risk is overreach — promising more than can be delivered.",
		"Capricorn":   "ASC Capricorn: the period begins with structural, disciplined energy. The nation focuses on government, institutions, and long-term building. Authority and responsibility dominate. The risk is rigidity — enforcing structure when flexibility is needed.",
		"Aquarius":    "ASC Aquarius: the period begins with innovative, collective energy. The nation focuses on reform, technology, and social movements. Congress and the people's voice dominate. The risk is disruption without direction — breaking things without building replacements.",
		"Pisces":      "ASC Pisces: the period begins with fluid, transcendent energy. The nation operates in a fog of idealism, propaganda, and collective emotion. The risk is delusion — believing the narrative rather than the facts.",
	}
	if m, ok := meanings[ascSign]; ok {
		return m
	}
	return fmt.Sprintf("ASC %s: the ingress period begins with %s energy shaping the national experience.", ascSign, strings.ToLower(ascSign))
}

// InterpretIngressMC returns an interpretation of the Midheaven in an
// ingress chart — the nation's public face and governmental direction.
func InterpretIngressMC(mcSign string, mcDegree float64) string {
	meanings := map[string]string{
		"Aries":       "MC Aries: the government's public face is initiatory and martial. The executive projects strength, decisiveness, and willingness to act. Military and executive authority are the visible face of power. The nation's reputation is tied to its capacity for action.",
		"Taurus":      "MC Taurus: the government's public face is steady and economic. The executive projects stability, material competence, and endurance. Economic management is the visible face of power. The nation's reputation is tied to its financial stewardship.",
		"Gemini":      "MC Gemini: the government's public face is communicative and adaptable. The executive projects intelligence, flexibility, and media savvy. Information management is the visible face of power. The nation's reputation is tied to its voice.",
		"Cancer":      "MC Cancer: the government's public face is protective and domestic. The executive projects care for the people, national security, and emotional connection. Public welfare is the visible face of power. The nation's reputation is tied to how it treats its own.",
		"Leo":         "MC Leo: the government's public face is commanding and performative. The executive projects confidence, authority, and star power. Personal leadership is the visible face of power. The nation's reputation is tied to its leader's image.",
		"Virgo":       "MC Virgo: the government's public face is competent and analytical. The executive projects efficiency, attention to detail, and problem-solving capacity. Administrative competence is the visible face of power. The nation's reputation is tied to its systems.",
		"Libra":       "MC Libra: the government's public face is diplomatic and balanced. The executive projects fairness, partnership, and international sophistication. Alliance management is the visible face of power. The nation's reputation is tied to its relationships.",
		"Scorpio":     "MC Scorpio: the government's public face is intense and strategic. The executive projects power, control, and the capacity for transformation. Intelligence and financial power are the visible face of authority. The nation's reputation is tied to its hidden strength.",
		"Sagittarius": "MC Sagittarius: the government's public face is expansive and ideological. The executive projects vision, moral authority, and global reach. The nation's beliefs are the visible face of power. The nation's reputation is tied to its principles.",
		"Capricorn":   "MC Capricorn: the government's public face is structural and authoritative. The executive projects discipline, responsibility, and institutional strength. Government itself is the visible face of power. The nation's reputation is tied to its institutions.",
		"Aquarius":    "MC Aquarius: the government's public face is innovative and collective. The executive projects reform, technological vision, and connection to the people. Democratic institutions are the visible face of power. The nation's reputation is tied to its capacity for change.",
		"Pisces":      "MC Pisces: the government's public face is elusive and idealistic. The executive projects vision, compassion, and spiritual authority — or confusion, scandal, and deception. The nation's reputation is tied to its myths and its mysteries.",
	}
	if m, ok := meanings[mcSign]; ok {
		return m
	}
	return fmt.Sprintf("MC %s: the government's public face is shaped by %s energy.", mcSign, strings.ToLower(mcSign))
}

// ── Mundane Pattern Interpretations ────────────────────────────────────

var mundanePatternMeanings = map[string]string{
	"T-Square":    "T-Square in the national chart: dynamic tension between three planetary forces. A pressure cooker that produces results through conflict. The nation is driven by an unresolved tension that demands action. The apex planet shows where the pressure must be released. This is not a comfortable configuration — it is a productive one. Nations with T-squares achieve through struggle what others achieve through ease.",
	"Grand Trine": "Grand Trine in the national chart: a closed loop of flowing trines between three planets. Effortless talent in one element — but talent that can become inertia. The nation has natural advantages that it may take for granted. The risk is complacency: when everything flows, nothing forces growth. The gift is a nation that can achieve with grace what others achieve with effort.",
	"Yod":         "Yod (Finger of God) in the national chart: two planets in sextile both quincunx a focal planet. A fated, uncomfortable assignment. The nation is being pointed toward something it cannot avoid. The focal planet carries a special mission — one the nation may resist but cannot escape. Yods in national charts often manifest as historical turning points that feel inevitable in retrospect.",
	"Grand Cross": "Grand Cross in the national chart: four planets in mutual square, two oppositions. Constant tension, extraordinary resilience. The nation is built to withstand pressure — it may not know peace, but it knows how to survive. This is the configuration of nations forged in crisis. The gift is unbreakable structure. The cost is that the structure never relaxes.",
	"Kite":        "Kite in the national chart: a Grand Trine with an opposition providing a release valve. The nation's natural talents have an outlet — a point of tension that channels the flowing energy into action. More dynamic than a pure Grand Trine. The opposition point shows where the nation's gifts must be applied. The risk is ignoring the release valve and coasting on talent.",
	"Cradle":      "Cradle in the national chart: an opposition supported by sextile-trine bridges on both sides. A supportive container for national tension. The opposition's conflict is held within a structure that makes it productive rather than destructive. The nation can hold opposing forces without being torn apart by them. The gift is the capacity to contain contradiction.",
	"Stellium":    "Stellium in the national chart: three or more planets concentrated in one sign or house. Intense focus of national energy in one area of life. The sign or house containing the stellium becomes the dominant force in national direction. The nation is a specialist, not a generalist. The gift is depth and intensity in one domain. The cost is that other domains are neglected or underdeveloped.",
	"Mystic Rectangle": "Mystic Rectangle in the national chart: two oppositions woven together by sextiles and trines. A rare harmonizing structure. The nation can hold opposing forces in creative tension — turning conflict into synthesis. This is the configuration of nations that transform contradiction into art, policy, or innovation. The gift is balance. The risk is that balance becomes paralysis.",
	"Wedge":      "Wedge in the national chart: a sextile, trine, and square forming a right triangle. Productive tension with a clear direction. The nation has a problem (the square) and the tools to solve it (the sextile and trine). More focused than a T-square, less passive than a Grand Trine. The square shows the work; the harmonious aspects show the resources.",
}

// ── Full Mundane Chart Interpretation ──────────────────────────────────

// MundaneChartInterpretation holds the full mundane interpretation.
type MundaneChartInterpretation struct {
	Name              string   `json:"name"`
	ChartType         string   `json:"chart_type"` // "natal", "ingress", "lunation", "eclipse"
	DateTime          string   `json:"date_time"`
	Location          string   `json:"location"`
	ASCSign           string   `json:"asc_sign"`
	MCSign            string   `json:"mc_sign"`
	ASCInterpretation string   `json:"asc_interpretation"`
	MCInterpretation  string   `json:"mc_interpretation"`
	PlanetHouses      []string `json:"planet_houses"`
	Patterns          []string `json:"patterns"`
	Summary           string   `json:"summary"`
}

// InterpretMundaneChartFull produces a complete mundane interpretation
// of a chart, using the collective-voice interpretive layer.
func InterpretMundaneChartFull(name, chartType string, chart *MundaneChart, orbDeg float64) *MundaneChartInterpretation {
	ascSign := signName(chart.ASC)
	mcSign := signName(chart.MC)

	report := &MundaneChartInterpretation{
		Name:              name,
		ChartType:         chartType,
		DateTime:          chart.Time.Format("2006-01-02 15:04:05 UTC"),
		Location:          fmt.Sprintf("%.2f, %.2f", chart.Lat, chart.Lon),
		ASCSign:           ascSign,
		MCSign:            mcSign,
		ASCInterpretation: InterpretIngressASC(ascSign, chart.ASC),
		MCInterpretation:  InterpretIngressMC(mcSign, chart.MC),
		PlanetHouses:      make([]string, 0),
		Patterns:          make([]string, 0),
	}

	// Planet-in-house interpretations
	houses := PlanetHouses(chart)
	var planetNames []string
	for n := range chart.Planets {
		planetNames = append(planetNames, n)
	}
	// Sort: classical planets first, then outer, then nodes
	order := map[string]int{
		"Sun": 0, "Moon": 1, "Mercury": 2, "Venus": 3, "Mars": 4,
		"Jupiter": 5, "Saturn": 6, "Uranus": 7, "Neptune": 8, "Pluto": 9,
		"Node": 10, "SouthNode": 11,
	}
	sort.Slice(planetNames, func(i, j int) bool {
		oi, oki := order[planetNames[i]]
		oj, okj := order[planetNames[j]]
		if oki && okj {
			return oi < oj
		}
		if oki {
			return true
		}
		if okj {
			return false
		}
		return planetNames[i] < planetNames[j]
	})

	for _, planet := range planetNames {
		house := houses[planet]
		if planetMap, ok := mundanePlanetHouseMeanings[planet]; ok {
			if meaning, ok := planetMap[house]; ok {
				report.PlanetHouses = append(report.PlanetHouses, meaning)
				continue
			}
		}
		// Fallback
		report.PlanetHouses = append(report.PlanetHouses,
			fmt.Sprintf("%s in house %d: %s expressed through %s.",
				planet, house, mundanePlanetMeanings[planet], mundaneHouseMeanings[house]))
	}

	// Patterns
	patterns := ChartPatterns(chart, orbDeg)
	for _, p := range patterns.Patterns {
		if meaning, ok := mundanePatternMeanings[p.Name]; ok {
			report.Patterns = append(report.Patterns,
				fmt.Sprintf("%s involving %s: %s", p.Name, strings.Join(p.Planets, ", "), meaning))
		} else {
			report.Patterns = append(report.Patterns,
				fmt.Sprintf("%s: %s", p.Name, strings.Join(p.Planets, ", ")))
		}
	}

	// Summary
	report.Summary = buildMundaneSummary(report, chart, orbDeg)

	return report
}

func buildMundaneSummary(report *MundaneChartInterpretation, chart *MundaneChart, orbDeg float64) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s chart cast for %s. ", report.ChartType, report.Name))
	sb.WriteString(fmt.Sprintf("ASC %s, MC %s. ", report.ASCSign, report.MCSign))

	// Count planets in angular houses (1, 4, 7, 10)
	houses := PlanetHouses(chart)
	angular := 0
	for _, h := range houses {
		if h == 1 || h == 4 || h == 7 || h == 10 {
			angular++
		}
	}
	if angular >= 4 {
		sb.WriteString(fmt.Sprintf("%d planets in angular houses — the chart is heavily action-oriented with national direction driven by visible, immediate forces. ", angular))
	} else if angular <= 1 {
		sb.WriteString("Few planets in angular houses — national direction operates through background forces rather than visible action. ")
	}

	// Patterns summary
	if len(report.Patterns) > 0 {
		sb.WriteString(fmt.Sprintf("%d significant planetary patterns detected. ", len(report.Patterns)))
	}

	// ASC/MC relationship
	ascElement := elementForSign(report.ASCSign)
	mcElement := elementForSign(report.MCSign)
	if ascElement == mcElement {
		sb.WriteString(fmt.Sprintf("The national persona and governmental direction share the %s element — identity and authority are aligned. ", ascElement))
	} else {
		sb.WriteString(fmt.Sprintf("The national persona (%s) and governmental direction (%s) operate in different elements — tension between how the nation is perceived and how it is governed. ", ascElement, mcElement))
	}

	return sb.String()
}

func elementForSign(sign string) string {
	switch sign {
	case "Aries", "Leo", "Sagittarius":
		return "fire"
	case "Taurus", "Virgo", "Capricorn":
		return "earth"
	case "Gemini", "Libra", "Aquarius":
		return "air"
	case "Cancer", "Scorpio", "Pisces":
		return "water"
	}
	return "unknown"
}

// ── Convenience: Full mundane interpretation from ingress event ────────

// InterpretIngressChartFull produces a full mundane interpretation of an
// ingress chart cast for a specific location.
func InterpretIngressChartFull(event IngressEvent, lat, lon float64, compute ComputeFunc, housesFunc HousesFunc) (*MundaneChartInterpretation, error) {
	chart, err := CastIngressChart(event, lat, lon, compute, housesFunc)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("%s Ingress %s", event.Planet, event.Sign)
	return InterpretMundaneChartFull(name, "ingress", chart, 5.0), nil
}

// ── Convenience: Full mundane interpretation from lunation event ───────

// InterpretLunationChartFull produces a full mundane interpretation of a
// lunation chart cast for a specific location.
func InterpretLunationChartFull(event LunationEvent, lat, lon float64, compute ComputeFunc, housesFunc HousesFunc) (*MundaneChartInterpretation, error) {
	chart, err := CastLunationChart(event, lat, lon, compute, housesFunc)
	if err != nil {
		return nil, err
	}
	return InterpretMundaneChartFull(event.Type, "lunation", chart, 5.0), nil
}

// ── Convenience: Full mundane interpretation of a national chart ───────

// InterpretNationalChartFull produces a full mundane interpretation of a
// national chart from the database.
func InterpretNationalChartFull(nationName string, compute ComputeFunc, housesFunc HousesFunc) (*MundaneChartInterpretation, error) {
	entry, ok := NationalChart(nationName)
	if !ok {
		return nil, fmt.Errorf("unknown nation: %s", nationName)
	}

	natalTime := time.Date(entry.Year, time.Month(entry.Month), entry.Day,
		int(entry.Hour), int((entry.Hour-float64(int(entry.Hour)))*60), 0, 0, time.UTC)
	chart, err := CastChart(natalTime, entry.Lat, entry.Lon, compute, housesFunc, 'W')
	if err != nil {
		return nil, fmt.Errorf("casting natal chart for %s: %w", nationName, err)
	}

	return InterpretMundaneChartFull(entry.Name, "natal", chart, 5.0), nil
}

// Ensure dignity import is used (for type compatibility)
var _ = dignity.InterpretChart
