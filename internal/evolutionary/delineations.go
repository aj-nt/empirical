package evolutionary

import (
	"fmt"
	"strings"
)

// ── Pluto in Sign ─────────────────────────────────────────────────────────

var plutoInSign = map[string]string{
	"Aries": "Pluto in Aries: the soul's evolutionary intent is to forge identity through direct action and raw courage. The death-rebirth cycle centers on self-assertion — each life chapter demands the old self be burned away so a more autonomous one can emerge. The shadow is impulsivity mistaken for transformation; the gift is the capacity to initiate without permission. The soul is learning that true power does not wait for consensus.",
	"Taurus": "Pluto in Taurus: the soul's evolutionary intent is to transform the relationship with substance, security, and the body. The death-rebirth cycle centers on attachment — what is owned, what is valued, what is held. Each evolutionary chapter demands release of material or sensory fixations that have become prisons. The shadow is hoarding what should be released; the gift is the capacity to build what endures. The soul is learning that security is not possession but presence.",
	"Gemini": "Pluto in Gemini: the soul's evolutionary intent is to transform the mind itself — how information is gathered, processed, and communicated. The death-rebirth cycle centers on belief systems and mental frameworks. Each evolutionary chapter demands the death of a worldview that has become too small. The shadow is intellectual restlessness substituting for depth; the gift is the capacity to hold multiple truths simultaneously. The soul is learning that the mind is a tool, not an identity.",
	"Cancer": "Pluto in Cancer: the soul's evolutionary intent is to transform the emotional foundation — family patterns, ancestral inheritance, and the definition of safety. The death-rebirth cycle centers on belonging. Each evolutionary chapter demands release of familial or tribal identifications that constrain the self. The shadow is emotional fusion mistaken for intimacy; the gift is the capacity to nurture without consuming. The soul is learning that home is internal, not external.",
	"Leo": "Pluto in Leo: the soul's evolutionary intent is to transform the relationship with creative power and self-expression. The death-rebirth cycle centers on recognition — who sees you, who doesn't, and whether that matters. Each evolutionary chapter demands the death of a performed self so a more authentic radiance can emerge. The shadow is needing the light reflected back to feel real; the gift is the capacity to create without audience. The soul is learning that radiance is inherent, not earned.",
	"Virgo": "Pluto in Virgo: the soul's evolutionary intent is to transform through craft, discernment, and the purification of daily life. The death-rebirth cycle centers on competence — what is mastered, what is improved, what is discarded. Each evolutionary chapter demands release of perfectionism that has become paralysis. The shadow is critique substituting for creation; the gift is the capacity to refine without destroying. The soul is learning that wholeness includes the imperfect.",
	"Libra": "Pluto in Libra: the soul's evolutionary intent is to transform through relationship — how the self is negotiated with the other. The death-rebirth cycle centers on partnership. Each evolutionary chapter demands the death of a relational pattern that has become codependent or self-erasing. The shadow is harmony at the cost of truth; the gift is the capacity to hold self and other in genuine equilibrium. The soul is learning that peace without honesty is not peace.",
	"Scorpio": "Pluto in Scorpio: the soul's evolutionary intent operates in its home sign — transformation is the native language. The death-rebirth cycle is not occasional but continuous. Each evolutionary chapter demands total surrender of what has outlived its purpose. The shadow is gripping what should be released, mistaking intensity for depth; the gift is the capacity to regenerate from absolute zero. The soul is learning that the phoenix does not rise from half-measures. At full strength: Pluto in domicile. The evolutionary pressure is relentless but the native is built for it.",
	"Sagittarius": "Pluto in Sagittarius: the soul's evolutionary intent is to transform belief systems, meaning structures, and the relationship with truth itself. The death-rebirth cycle centers on conviction — what is believed, what is preached, what is outgrown. Each evolutionary chapter demands the death of a dogma that has become a cage. The shadow is certainty substituting for wisdom; the gift is the capacity to hold faith without fundamentalism. The soul is learning that the map is not the territory.",
	"Capricorn": "Pluto in Capricorn: the soul's evolutionary intent is to transform the relationship with authority, structure, and ambition. The death-rebirth cycle centers on achievement — what is built, what is climbed, what is inherited from the father or the culture. Each evolutionary chapter demands release of a definition of success that was never the soul's own. The shadow is power mistaken for purpose; the gift is the capacity to build structures that serve rather than dominate. The soul is learning that the summit is not the point.",
	"Aquarius": "Pluto in Aquarius: the soul's evolutionary intent is to transform the relationship with the collective — how individuality is negotiated within the group. The death-rebirth cycle centers on belonging to something larger than the self. Each evolutionary chapter demands release of a group identity that has consumed the individual. The shadow is detachment substituting for freedom; the gift is the capacity to hold the self distinct while serving the whole. The soul is learning that true individuality does not require isolation.",
	"Pisces": "Pluto in Pisces: the soul's evolutionary intent is to transform through dissolution of boundaries — between self and other, self and cosmos, self and source. The death-rebirth cycle centers on surrender. Each evolutionary chapter demands release of the illusion of separation itself. The shadow is escapism mistaken for transcendence; the gift is the capacity to merge without losing the self. The soul is learning that the final transformation is the release of the need to transform.",
}

// ── Pluto in House ─────────────────────────────────────────────────────────

var plutoInHouse = map[int]string{
	1: "Pluto in the 1st house: the evolutionary work is worn on the surface. The soul's transformation is visible in the body, the presence, the first impression. Each life chapter reshapes identity itself — the native is not the same person they were. The death-rebirth cycle is public and undeniable. The shadow is over-identification with the persona's destruction and reconstruction; the gift is the capacity to reinvent without losing continuity. Angular: Pluto here makes the entire life a visible metamorphosis.",
	2: "Pluto in the 2nd house: the evolutionary work centers on resources, values, and self-worth. The soul transforms through what it owns and what it loses. Each life chapter demands release of a definition of worth tied to possession. The shadow is financial crisis as the only teacher; the gift is the capacity to rebuild from nothing and discover that value is intrinsic. The soul is learning that net worth and self-worth are not the same currency.",
	3: "Pluto in the 3rd house: the evolutionary work centers on the mind, communication, and the immediate environment. The soul transforms through what it learns and how it speaks. Each life chapter demands the death of a mental framework that has become too small. The shadow is intellectual obsession; the gift is the capacity to think beneath surfaces. The soul is learning that the most important conversations are the ones that change the speaker.",
	4: "Pluto in the 4th house: the evolutionary work centers on the foundation — family, home, ancestry, and the private self. The soul transforms through what it inherits and what it must release. Each life chapter demands the death of a familial pattern or a definition of safety. The shadow is being consumed by the past; the gift is the capacity to build a foundation that is truly one's own. Angular: Pluto here makes the private world the primary evolutionary theater.",
	5: "Pluto in the 5th house: the evolutionary work centers on creative expression, romance, and what the soul brings forth. The soul transforms through what it makes and who it loves. Each life chapter demands the death of a creative identity or a romantic pattern. The shadow is intensity consuming joy; the gift is the capacity to create from the underworld and bring back what others cannot reach. The soul is learning that the deepest art comes from what was survived.",
	6: "Pluto in the 6th house: the evolutionary work centers on work, health, service, and daily discipline. The soul transforms through what it does and how it cares for the body. Each life chapter demands the death of a work identity or a health pattern. The shadow is work as avoidance of the deeper work; the gift is the capacity to serve without losing the self. The soul is learning that purification is not punishment — it is preparation.",
	7: "Pluto in the 7th house: the evolutionary work centers on partnership. The soul transforms through the other — through who is chosen, who is lost, and what is mirrored back. Each life chapter demands the death of a relational pattern that has become a cage. The shadow is projecting the shadow onto the partner; the gift is the capacity to meet the other as equal without losing the self. Angular: Pluto here makes relationship the primary evolutionary crucible.",
	8: "Pluto in the 8th house: the evolutionary work operates in its natural house — the underworld is home. The soul transforms through intimacy, shared resources, death, and what is hidden. Each life chapter demands total surrender of control. The shadow is power struggles in merger; the gift is the capacity to navigate the depths without drowning. At full strength: Pluto in its natural house. The native is comfortable in crisis because crisis is where they grow.",
	9: "Pluto in the 9th house: the evolutionary work centers on belief, meaning, and the search for truth. The soul transforms through what it believes and what it outgrows. Each life chapter demands the death of a worldview. The shadow is certainty as armor; the gift is the capacity to hold faith after the old faith has burned. The soul is learning that the journey is the destination and the destination keeps moving.",
	10: "Pluto in the 10th house: the evolutionary work centers on career, reputation, and public role. The soul transforms through what it achieves and how it is seen. Each life chapter demands the death of a professional identity. The shadow is power for its own sake; the gift is the capacity to lead from the depths rather than the surface. Angular at the highest point: Pluto here makes the public life a visible death-rebirth cycle. The career is not a path — it is a series of metamorphoses.",
	11: "Pluto in the 11th house: the evolutionary work centers on community, collective ideals, and the future. The soul transforms through groups and what they demand. Each life chapter demands the death of a collective identity. The shadow is losing the self in the cause; the gift is the capacity to transform the group rather than be consumed by it. The soul is learning that the future is not inherited — it is forged.",
	12: "Pluto in the 12th house: the evolutionary work centers on the unconscious, the hidden, and what is surrendered to source. The soul transforms in solitude, in dreams, in what cannot be named. Each life chapter demands release of something the conscious mind cannot fully grasp. The shadow is dissolution without integration; the gift is the capacity to access the collective unconscious and return with wisdom. The soul is learning that the final transformation is the one that happens when no one is watching.",
}

// ── Pluto Polarity Point ────────────────────────────────────────────────────

var polarityPointInSign = map[string]string{
	"Aries": "The Pluto polarity point in Aries calls the soul toward direct self-assertion. The integration target is the courage to act without permission, to initiate without consensus, to be the first mover. The soul is learning that the power it has been developing in the Plutonian depths must eventually be wielded openly. The polarity demands the death of passivity.",
	"Taurus": "The Pluto polarity point in Taurus calls the soul toward embodiment and presence. The integration target is the capacity to be still, to inhabit the body fully, to find security in the senses rather than in control. The soul is learning that transformation must eventually land in the physical. The polarity demands the death of disembodiment.",
	"Gemini": "The Pluto polarity point in Gemini calls the soul toward curiosity and connection. The integration target is the capacity to hold multiple perspectives lightly, to communicate what has been learned in the depths, to translate the underworld into language. The soul is learning that wisdom must be shared to be complete. The polarity demands the death of isolation.",
	"Cancer": "The Pluto polarity point in Cancer calls the soul toward emotional belonging and nurturance. The integration target is the capacity to feel safe, to build home, to receive care without suspicion. The soul is learning that the fortress it built in the Plutonian depths must eventually open its doors. The polarity demands the death of emotional self-sufficiency as defense.",
	"Leo": "The Pluto polarity point in Leo calls the soul toward creative radiance and joyful self-expression. The integration target is the capacity to shine without permission, to create without audience, to be seen without armor. The soul is learning that the power developed in the depths must eventually come into the light. The polarity demands the death of hiding.",
	"Virgo": "The Pluto polarity point in Virgo calls the soul toward discernment, craft, and useful service. The integration target is the capacity to refine without destroying, to improve without obsessing, to serve without losing the self. The soul is learning that the depths must eventually produce something useful. The polarity demands the death of chaos for its own sake.",
	"Libra": "The Pluto polarity point in Libra calls the soul toward relationship and equilibrium. The integration target is the capacity to meet the other as equal, to hold harmony and truth simultaneously, to partner without merging. The soul is learning that the power developed in solitude must eventually be shared. The polarity demands the death of isolation as strength.",
	"Scorpio": "The Pluto polarity point in Scorpio calls the soul toward depth, merger, and total transformation. The integration target is the capacity to go all the way down, to merge without losing the self, to hold intensity without being consumed by it. The soul is learning that the surface life it has been cultivating must eventually yield to the depths. The polarity demands the death of superficiality.",
	"Sagittarius": "The Pluto polarity point in Sagittarius calls the soul toward meaning, expansion, and the search for truth. The integration target is the capacity to hold faith after certainty has burned, to explore without needing to arrive, to find meaning in the journey itself. The soul is learning that the depths must eventually produce a philosophy. The polarity demands the death of nihilism.",
	"Capricorn": "The Pluto polarity point in Capricorn calls the soul toward structure, authority, and earned achievement. The integration target is the capacity to build something that endures, to claim authority without domination, to achieve without losing the soul. The soul is learning that the power developed in the depths must eventually take form in the world. The polarity demands the death of formlessness.",
	"Aquarius": "The Pluto polarity point in Aquarius calls the soul toward collective contribution and radical individuality within the group. The integration target is the capacity to belong without conforming, to serve the future without losing the present, to hold the self distinct while connected. The soul is learning that the depths must eventually serve the collective. The polarity demands the death of isolation as identity.",
	"Pisces": "The Pluto polarity point in Pisces calls the soul toward surrender, compassion, and dissolution of the separate self. The integration target is the capacity to merge with source, to feel without boundary, to trust without evidence. The soul is learning that the power it has been developing must eventually be released. The polarity demands the death of control itself.",
}

// ── North Node in Sign ─────────────────────────────────────────────────────

var northNodeInSign = map[string]string{
	"Aries": "North Node in Aries: the growth direction is toward self-direction, courage, and the willingness to act alone. The soul is learning to initiate without permission, to assert without apology, to be the first mover. The past-life comfort zone is Libra's diplomacy and consensus-seeking — the familiar pattern of waiting for the other to validate the self. The evolutionary task is to discover that the self is sufficient, that action precedes certainty, and that the warrior's path is walked alone before it is walked with others.",
	"Taurus": "North Node in Taurus: the growth direction is toward embodiment, patience, and the cultivation of substance. The soul is learning to be still, to build slowly, to find security in the senses rather than in intensity. The past-life comfort zone is Scorpio's crisis and transformation — the familiar pattern of mistaking drama for depth. The evolutionary task is to discover that peace is not boredom, that stability is not stagnation, and that the body is not a vehicle to transcend but a home to inhabit.",
	"Gemini": "North Node in Gemini: the growth direction is toward curiosity, communication, and the willingness to hold multiple truths lightly. The soul is learning to gather information without needing it to cohere into a single grand narrative. The past-life comfort zone is Sagittarius's certainty and meaning-making — the familiar pattern of needing everything to mean something. The evolutionary task is to discover that not everything needs a thesis, that the question is sometimes more alive than the answer, and that the mind is a network, not a cathedral.",
	"Cancer": "North Node in Cancer: the growth direction is toward emotional belonging, nurturance, and the capacity to build home. The soul is learning to feel, to attach, to care for the vulnerable self and others. The past-life comfort zone is Capricorn's ambition and structural control — the familiar pattern of achieving instead of feeling. The evolutionary task is to discover that vulnerability is not weakness, that home is not a distraction from purpose, and that the heart has its own architecture.",
	"Leo": "North Node in Leo: the growth direction is toward creative radiance, joyful self-expression, and the courage to be seen. The soul is learning to shine without permission, to create without audience, to lead from the heart. The past-life comfort zone is Aquarius's detachment and collective orientation — the familiar pattern of serving the group at the expense of the self. The evolutionary task is to discover that individuality is not selfishness, that the spotlight is not dangerous, and that the self is worth the stage.",
	"Virgo": "North Node in Virgo: the growth direction is toward discernment, craft, and useful service. The soul is learning to refine, to improve, to bring order to chaos without losing compassion. The past-life comfort zone is Pisces's boundlessness and surrender — the familiar pattern of dissolving into the whole rather than doing the specific work. The evolutionary task is to discover that precision is not coldness, that service is not self-erasure, and that the details are where the divine hides.",
	"Libra": "North Node in Libra: the growth direction is toward relationship, harmony, and the capacity to meet the other as equal. The soul is learning to partner, to balance, to hold self and other in genuine equilibrium. The past-life comfort zone is Aries's solo action and self-assertion — the familiar pattern of going alone rather than negotiating. The evolutionary task is to discover that the other is not a threat to autonomy, that compromise is not weakness, and that the self is completed, not diminished, in relationship.",
	"Scorpio": "North Node in Scorpio: the growth direction is toward depth, transformation, and the capacity to go all the way down. The soul is learning to merge, to surrender control, to hold intensity without being consumed. The past-life comfort zone is Taurus's stability and sensory security — the familiar pattern of holding onto what is comfortable rather than facing what is true. The evolutionary task is to discover that safety is not the highest value, that the depths do not drown those who learn to swim, and that transformation requires the death of the familiar.",
	"Sagittarius": "North Node in Sagittarius: the growth direction is toward meaning, expansion, and the search for truth. The soul is learning to explore, to believe, to hold faith without fundamentalism. The past-life comfort zone is Gemini's information-gathering and mental multiplicity — the familiar pattern of collecting data without committing to meaning. The evolutionary task is to discover that not all truth is relative, that conviction is not ignorance, and that the journey requires a direction, not just movement.",
	"Capricorn": "North Node in Capricorn: the growth direction is toward structure, authority, and earned achievement. The soul is learning to build, to lead, to claim the father's seat without becoming the father's shadow. The past-life comfort zone is Cancer's emotional belonging and familial attachment — the familiar pattern of nesting rather than climbing. The evolutionary task is to discover that ambition is not betrayal of the heart, that authority is not domination, and that the summit is earned, not inherited.",
	"Aquarius": "North Node in Aquarius: the growth direction is toward collective contribution, radical individuality, and the capacity to serve the future. The soul is learning to belong without conforming, to innovate without isolating, to hold the self distinct while connected. The past-life comfort zone is Leo's personal radiance and creative self-expression — the familiar pattern of shining alone rather than distributing the light. The evolutionary task is to discover that the group does not consume the self, that the future is worth the present's sacrifice, and that individuality is most powerful when it serves something larger.",
	"Pisces": "North Node in Pisces: the growth direction is toward surrender, compassion, and dissolution of the separate self. The soul is learning to trust, to release, to merge with something larger than the ego. The past-life comfort zone is Virgo's discernment and control through precision — the familiar pattern of analyzing rather than accepting. The evolutionary task is to discover that not everything can be fixed, that surrender is not failure, and that the self is not diminished by dissolving — it is returned.",
}

// ── South Node in Sign ─────────────────────────────────────────────────────

var southNodeInSign = map[string]string{
	"Aries": "South Node in Aries: the past-life comfort zone is solo action, self-assertion, and the warrior's independence. The soul arrives with well-developed courage and initiative, but also with the habit of going alone. The familiar pattern is acting before consulting, asserting before listening, fighting before negotiating. The gift is the capacity to initiate without permission; the shadow is isolation mistaken for strength. The evolutionary task is to release the reflex to go it alone and learn the harder art of partnership.",
	"Taurus": "South Node in Taurus: the past-life comfort zone is stability, sensory security, and the accumulation of substance. The soul arrives with well-developed patience and the capacity to build, but also with the habit of holding on too long. The familiar pattern is mistaking comfort for safety, possession for worth, stillness for peace. The gift is the capacity to endure; the shadow is inertia mistaken for wisdom. The evolutionary task is to release the grip on the familiar and learn the harder art of transformation.",
	"Gemini": "South Node in Gemini: the past-life comfort zone is information, communication, and mental multiplicity. The soul arrives with well-developed curiosity and verbal skill, but also with the habit of collecting without committing. The familiar pattern is mistaking data for wisdom, conversation for connection, movement for progress. The gift is the capacity to hold multiple perspectives; the shadow is scattered attention mistaken for breadth. The evolutionary task is to release the reflex to gather more and learn the harder art of meaning-making.",
	"Cancer": "South Node in Cancer: the past-life comfort zone is emotional belonging, nurturance, and the private sanctuary. The soul arrives with well-developed empathy and the capacity to care, but also with the habit of nesting rather than climbing. The familiar pattern is mistaking safety for purpose, attachment for love, home for the whole world. The gift is the capacity to nurture; the shadow is emotional fusion mistaken for intimacy. The evolutionary task is to release the grip on the familiar hearth and learn the harder art of public achievement.",
	"Leo": "South Node in Leo: the past-life comfort zone is creative radiance, personal expression, and the need to be seen. The soul arrives with well-developed charisma and the capacity to shine, but also with the habit of needing the light reflected back. The familiar pattern is mistaking attention for love, performance for presence, the stage for the world. The gift is the capacity to create; the shadow is the audience becoming the point. The evolutionary task is to release the need for recognition and learn the harder art of collective contribution.",
	"Virgo": "South Node in Virgo: the past-life comfort zone is discernment, craft, and the improvement of everything within reach. The soul arrives with well-developed analytical skill and the capacity to refine, but also with the habit of critique substituting for acceptance. The familiar pattern is mistaking perfection for completion, analysis for understanding, the part for the whole. The gift is the capacity to improve; the shadow is the inability to leave anything unoptimized. The evolutionary task is to release the need to fix everything and learn the harder art of surrender.",
	"Libra": "South Node in Libra: the past-life comfort zone is relationship, harmony, and the negotiation of self with other. The soul arrives with well-developed diplomacy and the capacity to partner, but also with the habit of waiting for consensus before acting. The familiar pattern is mistaking peace for truth, compromise for resolution, the other's validation for self-knowledge. The gift is the capacity to harmonize; the shadow is self-erasure mistaken for grace. The evolutionary task is to release the reflex to check with the other and learn the harder art of solo action.",
	"Scorpio": "South Node in Scorpio: the past-life comfort zone is depth, intensity, and the navigation of crisis. The soul arrives with well-developed psychological penetration and the capacity to survive, but also with the habit of mistaking drama for significance. The familiar pattern is gripping what should be released, controlling what should be trusted, diving when floating would suffice. The gift is the capacity to transform; the shadow is intensity mistaken for depth. The evolutionary task is to release the addiction to crisis and learn the harder art of peace.",
	"Sagittarius": "South Node in Sagittarius: the past-life comfort zone is meaning, expansion, and the search for truth. The soul arrives with well-developed faith and the capacity to explore, but also with the habit of mistaking certainty for wisdom. The familiar pattern is preaching when listening would serve, expanding when integrating would serve, moving when staying would serve. The gift is the capacity to believe; the shadow is dogma mistaken for truth. The evolutionary task is to release the need for everything to mean something and learn the harder art of holding questions without answers.",
	"Capricorn": "South Node in Capricorn: the past-life comfort zone is structure, authority, and earned achievement. The soul arrives with well-developed discipline and the capacity to build, but also with the habit of mistaking position for worth. The familiar pattern is climbing when nesting would serve, controlling when feeling would serve, achieving when being would serve. The gift is the capacity to lead; the shadow is the summit becoming the only point. The evolutionary task is to release the identification with achievement and learn the harder art of emotional belonging.",
	"Aquarius": "South Node in Aquarius: the past-life comfort zone is collective contribution, radical individuality, and service to the future. The soul arrives with well-developed vision and the capacity to innovate, but also with the habit of mistaking detachment for freedom. The familiar pattern is serving the group at the expense of the self, the cause at the expense of the heart, the future at the expense of the present. The gift is the capacity to see the system; the shadow is the personal becoming invisible. The evolutionary task is to release the identification with the collective and learn the harder art of personal radiance.",
	"Pisces": "South Node in Pisces: the past-life comfort zone is surrender, compassion, and dissolution of boundaries. The soul arrives with well-developed empathy and the capacity to merge, but also with the habit of mistaking escape for transcendence. The familiar pattern is dissolving when defining would serve, accepting when discerning would serve, trusting when questioning would serve. The gift is the capacity to feel the whole; the shadow is the self lost in the ocean. The evolutionary task is to release the reflex to dissolve and learn the harder art of precision and boundaries.",
}

// ── Saturn in Sign ─────────────────────────────────────────────────────────

var saturnInSign = map[string]string{
	"Aries": "Saturn in Aries: the gatekeeper tests through the impulse to act. The soul's curriculum is patience in the face of urgency, discipline in the channeling of raw will. Saturn here demands that initiative be earned — the native cannot simply charge forward; they must learn when to advance and when to wait. The shadow is frustration becoming resignation; the gift is the capacity to act with precision after the pause. The soul is learning that the warrior who masters timing masters the battle.",
	"Taurus": "Saturn in Taurus: the gatekeeper tests through substance and security. The soul's curriculum is the relationship with material reality — what is owned, what is valued, what endures. Saturn here demands that worth be built slowly, that patience be practiced until it becomes nature. The shadow is scarcity thinking becoming a self-fulfilling prophecy; the gift is the capacity to build what cannot be taken. The soul is learning that true security is not what is held but what has been earned.",
	"Gemini": "Saturn in Gemini: the gatekeeper tests through the mind. The soul's curriculum is the discipline of thought — learning to focus, to finish, to go deep rather than wide. Saturn here demands that communication be precise, that knowledge be earned through study rather than gathered through curiosity alone. The shadow is mental restlessness becoming superficiality; the gift is the capacity to think with rigor. The soul is learning that the mind is a muscle that must be trained, not just fed.",
	"Cancer": "Saturn in Cancer: the gatekeeper tests through emotional foundation. The soul's curriculum is the architecture of feeling — learning to hold emotion without being consumed, to nurture without losing the self, to build home that is truly one's own. Saturn here demands that emotional maturity be earned through facing what was inherited. The shadow is emotional withholding becoming isolation; the gift is the capacity to feel deeply while standing firmly. The soul is learning that the strongest walls are built around the softest centers.",
	"Leo": "Saturn in Leo: the gatekeeper tests through creative expression and the need to be seen. The soul's curriculum is the discipline of radiance — learning to shine without audience, to create without validation, to lead without needing followers. Saturn here demands that self-expression be earned through mastery of craft, not granted by charisma alone. The shadow is the fear of invisibility becoming self-diminishment; the gift is the capacity to create from substance rather than performance. The soul is learning that the sun does not need permission to rise.",
	"Virgo": "Saturn in Virgo: the gatekeeper tests through discernment and craft. The soul's curriculum is the discipline of precision — learning to refine without destroying, to improve without obsessing, to serve without losing the self. Saturn here demands that competence be earned through practice, not claimed through critique. The shadow is perfectionism becoming paralysis; the gift is the capacity to produce work of genuine mastery. The soul is learning that the master has failed more times than the apprentice has tried.",
	"Libra": "Saturn in Libra: the gatekeeper tests through relationship. The soul's curriculum is the discipline of partnership — learning to hold self and other in genuine equilibrium, to commit without losing autonomy, to harmonize without erasing truth. Saturn here demands that relationship be earned through maturity, not entered through need. The shadow is commitment-phobia becoming isolation; the gift is the capacity to build partnership that endures. The soul is learning that the strongest bonds are forged, not found.",
	"Scorpio": "Saturn in Scorpio: the gatekeeper tests through depth and control. The soul's curriculum is the discipline of surrender — learning to release what cannot be controlled, to trust what cannot be verified, to face the underworld without becoming its permanent resident. Saturn here demands that power be earned through integrity, not seized through manipulation. The shadow is control becoming prison; the gift is the capacity to hold power without being corrupted by it. The soul is learning that the only thing worth controlling is the self.",
	"Sagittarius": "Saturn in Sagittarius: the gatekeeper tests through belief and meaning. The soul's curriculum is the discipline of faith — learning to hold conviction without dogma, to explore without escaping, to find meaning that survives doubt. Saturn here demands that wisdom be earned through experience, not adopted through ideology. The shadow is certainty becoming fundamentalism; the gift is the capacity to believe after everything has been questioned. The soul is learning that the strongest faith is the one that has survived its own demolition.",
	"Capricorn": "Saturn in Capricorn: the gatekeeper operates in its home sign — the curriculum is the architecture of a life. Saturn here demands that everything be earned: authority, achievement, respect. The soul's test is ambition itself — learning to climb without losing the self, to build without becoming the structure, to lead without becoming the father's shadow. The shadow is achievement becoming the only measure of worth; the gift is the capacity to build what endures beyond the self. At full strength: Saturn in domicile. The gatekeeper is not cruel here — it is exact. What is earned is real.",
	"Aquarius": "Saturn in Aquarius: the gatekeeper tests through collective contribution and individuality within the group. The soul's curriculum is the discipline of belonging — learning to serve the future without losing the present, to hold the self distinct while connected, to innovate without isolating. Saturn here demands that vision be earned through understanding of systems, not claimed through rebellion alone. The shadow is detachment becoming alienation; the gift is the capacity to build structures that liberate rather than constrain. The soul is learning that the most radical act is to build something that lasts.",
	"Pisces": "Saturn in Pisces: the gatekeeper tests through surrender and boundaries. The soul's curriculum is the discipline of dissolution — learning to release without escaping, to feel without drowning, to trust without naivety. Saturn here demands that compassion be earned through facing suffering, not adopted through sentiment. The shadow is avoidance becoming a life unlived; the gift is the capacity to hold form and formlessness simultaneously. The soul is learning that the strongest container is the one that knows when to open.",
}

// ── Saturn in House ────────────────────────────────────────────────────────

var saturnInHouse = map[int]string{
	1: "Saturn in the 1st house: the gatekeeper stands at the threshold of identity. The soul's curriculum is the self itself — learning to inhabit the body, to claim presence, to be seen without armor. Saturn here demands that identity be built, not inherited. The shadow is self-diminishment becoming invisibility; the gift is the capacity to develop a self of genuine substance. Angular: Saturn here makes the entire life a project of becoming.",
	2: "Saturn in the 2nd house: the gatekeeper tests through resources and self-worth. The soul's curriculum is the relationship with value — learning to earn, to keep, to know what the self is worth independent of what it owns. Saturn here demands that financial maturity be earned through discipline. The shadow is scarcity consciousness becoming self-fulfilling poverty; the gift is the capacity to build lasting material security.",
	3: "Saturn in the 3rd house: the gatekeeper tests through communication and learning. The soul's curriculum is the discipline of the mind — learning to speak with precision, to listen with depth, to think with rigor. Saturn here demands that knowledge be earned through study, not gathered through curiosity. The shadow is intellectual insecurity becoming silence; the gift is the capacity to communicate with authority earned through mastery.",
	4: "Saturn in the 4th house: the gatekeeper tests through the foundation. The soul's curriculum is home, family, and the private self — learning to build sanctuary, to face what was inherited, to become the parent to the self that was never had. Saturn here demands that emotional security be built from the ground up. The shadow is the past becoming a prison; the gift is the capacity to build a foundation that is truly one's own. Angular: Saturn here makes the private world the primary site of the soul's work.",
	5: "Saturn in the 5th house: the gatekeeper tests through creativity and joy. The soul's curriculum is the discipline of expression — learning to create without audience, to play without guilt, to love without fear. Saturn here demands that creative mastery be earned through practice, not granted by talent. The shadow is the fear of judgment becoming creative paralysis; the gift is the capacity to produce work of disciplined joy.",
	6: "Saturn in the 6th house: the gatekeeper tests through work, health, and daily discipline. The soul's curriculum is the architecture of the ordinary — learning to serve without losing the self, to care for the body without obsession, to work without becoming the work. Saturn here demands that competence be earned through repetition. The shadow is work becoming avoidance of the deeper work; the gift is the capacity to build a life of meaningful service.",
	7: "Saturn in the 7th house: the gatekeeper tests through partnership. The soul's curriculum is the discipline of the other — learning to commit, to stay, to hold self and other in genuine equilibrium. Saturn here demands that relationship be earned through maturity, not entered through need. The shadow is fear of intimacy becoming isolation; the gift is the capacity to build partnership that deepens with time. Angular: Saturn here makes relationship the primary site of the soul's maturation.",
	8: "Saturn in the 8th house: the gatekeeper tests through depth, intimacy, and shared resources. The soul's curriculum is the discipline of merger — learning to trust, to share, to face the underworld without becoming its permanent resident. Saturn here demands that power be earned through integrity. The shadow is control becoming isolation; the gift is the capacity to hold intimacy without being consumed by it.",
	9: "Saturn in the 9th house: the gatekeeper tests through belief and meaning. The soul's curriculum is the discipline of faith — learning to hold conviction without dogma, to explore without escaping, to find truth that survives doubt. Saturn here demands that wisdom be earned through experience. The shadow is certainty becoming fundamentalism; the gift is the capacity to believe after everything has been questioned.",
	10: "Saturn in the 10th house: the gatekeeper tests through career, reputation, and public role. The soul's curriculum is the discipline of authority — learning to lead, to build, to claim the father's seat without becoming the father's shadow. Saturn here demands that achievement be earned through sustained effort. The shadow is ambition becoming the only measure of worth; the gift is the capacity to build a legacy. Angular at the highest point: Saturn here makes the public life the primary site of the soul's maturation.",
	11: "Saturn in the 11th house: the gatekeeper tests through community and collective ideals. The soul's curriculum is the discipline of belonging — learning to serve the future without losing the present, to hold the self distinct while connected. Saturn here demands that contribution be earned through understanding of systems. The shadow is detachment becoming alienation; the gift is the capacity to build structures that serve the collective.",
	12: "Saturn in the 12th house: the gatekeeper tests through the hidden, the unconscious, and what is surrendered. The soul's curriculum is the discipline of solitude — learning to face the self without distraction, to hold boundaries without walls, to serve without recognition. Saturn here demands that spiritual maturity be earned through genuine inner work. The shadow is isolation becoming a life unlived; the gift is the capacity to access depths that the surface world cannot reach.",
}

// ── Skipped Step Delineations ──────────────────────────────────────────────

var skippedStepPlanet = map[string]string{
	"Sun": "The Sun in hard aspect to the nodal axis: the core identity itself is unfinished business. The soul arrives with an unresolved relationship to self-expression, authority, and the right to shine. The familiar pattern is either over-identification with the ego or its suppression — the soul either dominated past-life contexts or was eclipsed by them. The evolutionary task is to integrate the Sun's radiance consciously: to lead without dominating, to shine without needing the light reflected back, to claim the self without apology. This is not a minor adjustment — it is the renegotiation of identity itself.",
	"Moon": "The Moon in hard aspect to the nodal axis: the emotional body carries unresolved karma. The soul arrives with an inherited emotional pattern — a way of feeling, attaching, and seeking safety — that was adaptive in past-life contexts but now constrains growth. The familiar pattern is either emotional fusion or emotional withholding. The evolutionary task is to feel without being consumed, to attach without losing the self, to find security that is internal rather than external. The Moon's skipped step is felt in the body before it is understood by the mind.",
	"Mercury": "Mercury in hard aspect to the nodal axis: the mind itself is the site of unfinished business. The soul arrives with an unresolved relationship to communication, knowledge, and truth-telling. The familiar pattern is either intellectual dominance or intellectual silence — the soul either used words as weapons in past-life contexts or was silenced by them. The evolutionary task is to speak with precision and integrity, to listen as deeply as one speaks, to use the mind as a bridge rather than a barrier.",
	"Venus": "Venus in hard aspect to the nodal axis: the capacity to love, value, and receive is unfinished business. The soul arrives with an unresolved relationship to worth, beauty, and partnership. The familiar pattern is either over-giving to earn love or withholding to avoid vulnerability. The evolutionary task is to love without losing the self, to receive without debt, to know one's worth independent of the other's valuation. Venus skipped steps often manifest as repeating relational patterns that the soul recognizes but cannot yet break.",
	"Mars": "Mars in hard aspect to the nodal axis: the will to act, to assert, to fight is unfinished business. The soul arrives with an unresolved relationship to anger, desire, and self-assertion. The familiar pattern is either aggression without cause or passivity mistaken for peace — the soul either wielded force destructively in past-life contexts or was disempowered by it. The evolutionary task is to act with precision, to assert without domination, to channel the warrior's fire into creation rather than destruction.",
	"Jupiter": "Jupiter in hard aspect to the nodal axis: the relationship to faith, expansion, and meaning is unfinished business. The soul arrives with an unresolved relationship to belief — either dogmatic certainty or nihilistic doubt. The familiar pattern is either imposing truth on others or refusing to commit to any truth at all. The evolutionary task is to hold faith without fundamentalism, to expand without escaping, to find meaning that survives honest questioning. Jupiter skipped steps often manifest as a pendulum between grandiosity and collapse.",
	"Saturn": "Saturn in hard aspect to the nodal axis: the relationship to authority, structure, and discipline is unfinished business. The soul arrives with an unresolved relationship to power — either submission to external authority or rebellion against all structure. The familiar pattern is either over-responsibility that crushes the self or under-responsibility that avoids maturity. The evolutionary task is to claim authority over the self without becoming the tyrant one fled, to build structures that serve rather than imprison, to earn mastery without becoming the master one resented.",
	"Uranus": "Uranus in hard aspect to the nodal axis: the relationship to freedom, individuality, and disruption is unfinished business. The soul arrives with an unresolved relationship to belonging — either conformity that erased the self or rebellion that isolated it. The familiar pattern is either suppressing uniqueness to fit in or rejecting all connection to stay free. The evolutionary task is to hold individuality and connection simultaneously, to innovate without destroying, to belong without conforming. Uranus skipped steps often manifest as a life that oscillates between radical breaks and desperate attempts to fit in.",
	"Neptune": "Neptune in hard aspect to the nodal axis: the relationship to surrender, illusion, and transcendence is unfinished business. The soul arrives with an unresolved relationship to reality itself — either escapism that avoided the world or disillusionment that rejected the transcendent. The familiar pattern is either dissolving boundaries to the point of self-loss or building walls to the point of isolation. The evolutionary task is to hold vision and ground simultaneously, to trust without naivety, to surrender without escaping. Neptune skipped steps are the most subtle and the most disorienting — the soul must learn to see clearly while keeping the capacity for wonder.",
	"Pluto": "Pluto in hard aspect to the nodal axis: the relationship to power, control, and transformation is unfinished business — and this is the deepest skipped step possible. The soul arrives with an unresolved relationship to the death-rebirth cycle itself. The familiar pattern is either wielding power over others or being destroyed by it. The evolutionary task is to transform without destroying, to hold power without corruption, to face the underworld without becoming its permanent resident. Pluto skipped steps are not subtle — they manifest as life-or-death crises that force the soul to evolve or perish. This is the skipped step that underlies all others.",
}

// ── Narrative Builder ──────────────────────────────────────────────────────

func buildNarrative(r *EvolutionaryReport) string {
	var b strings.Builder

	// Title
	b.WriteString(fmt.Sprintf("Evolutionary Astrology Reading for %s\n\n", r.Name))

	// ── Pluto ──
	b.WriteString("═══ PLUTO — THE EVOLUTIONARY JOURNEY ═══\n\n")
	b.WriteString(fmt.Sprintf("Pluto in %s, %s.\n\n", r.Pluto.Sign, houseLabel(r.Pluto.House)))

	if d, ok := plutoInSign[r.Pluto.Sign]; ok {
		b.WriteString(d)
		b.WriteString("\n\n")
	}
	if d, ok := plutoInHouse[r.Pluto.House]; ok {
		b.WriteString(d)
		b.WriteString("\n\n")
	}

	// Polarity point
	b.WriteString(fmt.Sprintf("Pluto Polarity Point: %s, %s.\n\n",
		r.PlutoPolarity.Sign, houseLabel(r.PlutoPolarity.House)))
	if d, ok := polarityPointInSign[r.PlutoPolarity.Sign]; ok {
		b.WriteString(d)
		b.WriteString("\n\n")
	}

	// ── Nodal Axis ──
	b.WriteString("═══ NODAL AXIS — THE KARMIC DIRECTION ═══\n\n")
	b.WriteString(fmt.Sprintf("North Node in %s, %s.\n", r.NorthNode.Sign, houseLabel(r.NorthNode.House)))
	b.WriteString(fmt.Sprintf("South Node in %s, %s.\n\n", r.SouthNode.Sign, houseLabel(r.SouthNode.House)))

	if d, ok := northNodeInSign[r.NorthNode.Sign]; ok {
		b.WriteString(d)
		b.WriteString("\n\n")
	}
	if d, ok := southNodeInSign[r.SouthNode.Sign]; ok {
		b.WriteString(d)
		b.WriteString("\n\n")
	}

	// South Node ruler
	b.WriteString(fmt.Sprintf("South Node Ruler: %s in %s, %s.\n\n",
		r.SouthNodeRuler.Planet, r.SouthNodeRuler.Sign, houseLabel(r.SouthNodeRuler.House)))
	b.WriteString(fmt.Sprintf("The South Node ruler shows the central drama of the past-life pattern. "+
		"With the ruler in %s, %s, the soul's familiar operating system is %s expression through %s matters. "+
		"This is the energy that must be consciously redirected toward the North Node's growth direction.\n\n",
		r.SouthNodeRuler.Sign, houseLabel(r.SouthNodeRuler.House),
		r.SouthNodeRuler.Planet, houseLabel(r.SouthNodeRuler.House)))

	// ── Saturn ──
	b.WriteString("═══ SATURN — THE GATEKEEPER ═══\n\n")
	b.WriteString(fmt.Sprintf("Saturn in %s, %s.\n\n", r.Saturn.Sign, houseLabel(r.Saturn.House)))

	if d, ok := saturnInSign[r.Saturn.Sign]; ok {
		b.WriteString(d)
		b.WriteString("\n\n")
	}
	if d, ok := saturnInHouse[r.Saturn.House]; ok {
		b.WriteString(d)
		b.WriteString("\n\n")
	}

	// ── Skipped Steps ──
	if len(r.SkippedSteps) > 0 {
		b.WriteString("═══ SKIPPED STEPS — UNFINISHED BUSINESS ═══\n\n")
		for _, ss := range r.SkippedSteps {
			b.WriteString(fmt.Sprintf("%s %s the nodal axis (orb %.2f°).\n\n",
				ss.Planet, ss.Aspect, ss.Orb))
			if d, ok := skippedStepPlanet[ss.Planet]; ok {
				b.WriteString(d)
				b.WriteString("\n\n")
			}
		}
	}

	// ── Synthesis ──
	b.WriteString("═══ SYNTHESIS ═══\n\n")
	b.WriteString(buildSynthesis(r))

	return b.String()
}

func buildSynthesis(r *EvolutionaryReport) string {
	var b strings.Builder

	// Opening thread: Pluto + polarity point
	b.WriteString(fmt.Sprintf(
		"The evolutionary path centers on Pluto in %s (%s), with the integration target at the polarity point in %s (%s). ",
		r.Pluto.Sign, houseLabel(r.Pluto.House),
		r.PlutoPolarity.Sign, houseLabel(r.PlutoPolarity.House)))

	// Nodal direction
	b.WriteString(fmt.Sprintf(
		"The nodal axis runs from the South Node in %s (%s) to the North Node in %s (%s), "+
			"showing the karmic direction from %s patterns toward %s growth. ",
		r.SouthNode.Sign, houseLabel(r.SouthNode.House),
		r.NorthNode.Sign, houseLabel(r.NorthNode.House),
		r.SouthNode.Sign, r.NorthNode.Sign))

	// Saturn gatekeeper
	b.WriteString(fmt.Sprintf(
		"Saturn in %s (%s) stands as the gatekeeper, testing readiness through %s discipline and %s experience. ",
		r.Saturn.Sign, houseLabel(r.Saturn.House),
		r.Saturn.Sign, houseLabel(r.Saturn.House)))

	// Skipped steps integration
	if len(r.SkippedSteps) > 0 {
		planets := make([]string, len(r.SkippedSteps))
		for i, ss := range r.SkippedSteps {
			planets[i] = ss.Planet
		}
		b.WriteString(fmt.Sprintf(
			"The skipped step(s) involving %s represent unfinished business that must be consciously integrated. ",
			strings.Join(planets, " and ")))
		b.WriteString("These are not obstacles to the path — they are the path. ")
	}

	// Closing
	b.WriteString(fmt.Sprintf(
		"The soul's work in this life is to transform %s patterns through %s experience, "+
			"moving from %s toward %s, with Saturn in %s ensuring that nothing is claimed that has not been earned.",
		r.Pluto.Sign, houseLabel(r.Pluto.House),
		r.SouthNode.Sign, r.NorthNode.Sign,
		r.Saturn.Sign))

	return b.String()
}

func houseLabel(h int) string {
	labels := map[int]string{
		1: "1st house", 2: "2nd house", 3: "3rd house",
		4: "4th house", 5: "5th house", 6: "6th house",
		7: "7th house", 8: "8th house", 9: "9th house",
		10: "10th house", 11: "11th house", 12: "12th house",
	}
	if l, ok := labels[h]; ok {
		return l
	}
	return fmt.Sprintf("house %d", h)
}
