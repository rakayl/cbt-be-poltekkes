-- =============================================================================
-- Migration 000006: Seed Authentic Question Banks for TOEFL ITP & TOEIC L&R
-- Schema: cat
-- Using Dynamic Question Model:
--   cat.cat_bank_stimulus -> cat.cat_stimulus_items -> cat.cat_item_options
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- SECTION 1: TOEFL ITP - LISTENING COMPREHENSION (kodesoal: 'TOEFL_L_01')
-- ─────────────────────────────────────────────────────────────────────────────

-- Stimulus 1: Part A - Short Conversation 1
INSERT INTO cat.cat_bank_stimulus (kodesoal, stimulus_code, title, narrative_text, media_type, media_url, category_tag, difficulty_level, sort_order)
VALUES (
    'TOEFL_L_01', 'TOEFL_L_A01', 'Short Conversation: Biology Lab Assignment',
    '(Man): Did you finish the lab report for Dr. Peterson''s biology class yet?\n(Woman): Not yet. I had trouble analyzing the second batch of microscope slides, so I''m planning to consult the teaching assistant tomorrow morning.\n(Narrator): What does the woman imply?',
    'AUDIO', 'https://actions.google.com/sounds/v1/human_voices/human_voice_female_dialogue.ogg',
    'TOEFL_LISTENING_PART_A', 2, 1
) ON CONFLICT DO NOTHING;

INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 1, 'What will the woman most likely do tomorrow morning?', 'SINGLE_CHOICE', 1.0, 'C', 'The woman states she will consult the teaching assistant because she had trouble analyzing the slides.'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEFL_L_A01';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'Submit her final biology lab report to Dr. Peterson.', 1),
    ('B', 'Prepare a new batch of microscope slides in the laboratory.', 2),
    ('C', 'Ask the teaching assistant for help with the assignment.', 3),
    ('D', 'Cancel her appointment with the biology professor.', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEFL_L_A01';

-- Stimulus 2: Part A - Short Conversation 2
INSERT INTO cat.cat_bank_stimulus (kodesoal, stimulus_code, title, narrative_text, media_type, media_url, category_tag, difficulty_level, sort_order)
VALUES (
    'TOEFL_L_01', 'TOEFL_L_A02', 'Short Conversation: Library Books Reservation',
    '(Woman): Are the reference textbooks on medical pharmacology on reserve, or can we check them out overnight?\n(Man): They are strictly for library use only, but you can photocopy any chapters you need on the third floor.\n(Narrator): What does the man mean?',
    'AUDIO', 'https://actions.google.com/sounds/v1/human_voices/human_voice_male_dialogue.ogg',
    'TOEFL_LISTENING_PART_A', 2, 2
) ON CONFLICT DO NOTHING;

INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 1, 'What does the man indicate about the pharmacology textbooks?', 'SINGLE_CHOICE', 1.0, 'B', 'The man states that the books are strictly for library use only, meaning they cannot be removed from the library.'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEFL_L_A02';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'Students may borrow the books for one week.', 1),
    ('B', 'The textbooks cannot be taken out of the library building.', 2),
    ('C', 'The library has run out of medical reference books.', 3),
    ('D', 'Photocopy machines are currently out of service.', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEFL_L_A02';

-- Stimulus 3: Part B - Longer Conversation (Campus Healthcare System)
INSERT INTO cat.cat_bank_stimulus (kodesoal, stimulus_code, title, narrative_text, media_type, media_url, category_tag, difficulty_level, sort_order)
VALUES (
    'TOEFL_L_01', 'TOEFL_L_B01', 'Long Conversation: Student Health Center Insurance Plan',
    '(Man): Hi Sandra, are you heading over to the campus clinic for your annual health screening?\n(Woman): Yes, Mark. Since the university updated its mandatory health coverage this semester, all international and nursing students must complete a physical examination and update their immunization records before clinical rotations begin next month.\n(Man): Oh, I see. Does our student insurance plan cover the required hepatitis and tetanus booster vaccinations, or do we have to pay out of pocket?\n(Woman): Everything listed on the university health portal is 100% subsidized under the comprehensive student health fee, as long as you visit the on-campus clinic during regular operating hours.\n(Man): That is a relief! I better schedule my appointment before the clinic schedule fills up.',
    'AUDIO', 'https://actions.google.com/sounds/v1/human_voices/human_voice_female_dialogue.ogg',
    'TOEFL_LISTENING_PART_B', 2, 3
) ON CONFLICT DO NOTHING;

-- Question 1 under Stimulus 3
INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 1, 'Why is the woman visiting the student health center?', 'SINGLE_CHOICE', 1.0, 'A', 'She needs to complete a physical examination and update immunization records before clinical rotations.'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEFL_L_B01';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'To fulfill health screening requirements for upcoming clinical practice.', 1),
    ('B', 'To pick up prescription medication for a sudden illness.', 2),
    ('C', 'To apply for a part-time job as a clinical assistant.', 3),
    ('D', 'To request a refund for her university tuition fees.', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEFL_L_B01' AND i.item_order = 1;

-- Question 2 under Stimulus 3
INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 2, 'What does the woman say regarding vaccination costs?', 'SINGLE_CHOICE', 1.0, 'D', 'The required vaccinations are fully subsidized under the student health insurance fee.'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEFL_L_B01';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'Students must pay an additional deductible at the counter.', 1),
    ('B', 'Vaccines are only covered for senior medical residents.', 2),
    ('C', 'Costs are reimbursed after submitting an insurance claim form.', 3),
    ('D', 'The required vaccinations are completely covered by the student health fee.', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEFL_L_B01' AND i.item_order = 2;

-- ─────────────────────────────────────────────────────────────────────────────
-- SECTION 2: TOEFL ITP - STRUCTURE & WRITTEN EXPRESSION (kodesoal: 'TOEFL_S_01')
-- ─────────────────────────────────────────────────────────────────────────────

-- Question 1: Incomplete Sentence (Subject-Verb Agreement / Clause)
INSERT INTO cat.cat_bank_stimulus (kodesoal, stimulus_code, title, narrative_text, media_type, category_tag, difficulty_level, sort_order)
VALUES (
    'TOEFL_S_01', 'TOEFL_S_01', 'Structure Question 1',
    'Choose the word or phrase that best completes the sentence.',
    'NONE', 'TOEFL_STRUCTURE_PART_A', 2, 1
) ON CONFLICT DO NOTHING;

INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 1, 'Not only ________ significant amounts of vitamin C, but fresh oranges also provide essential dietary fiber.', 'SINGLE_CHOICE', 1.0, 'B', 'Inversion after negative adverbial phrase "Not only" requires auxiliary verb + subject: "do they contain".'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEFL_S_01';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'they contain', 1),
    ('B', 'do they contain', 2),
    ('C', 'they are containing', 3),
    ('D', 'contain they', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEFL_S_01';

-- Question 2: Incomplete Sentence (Participle Clause)
INSERT INTO cat.cat_bank_stimulus (kodesoal, stimulus_code, title, narrative_text, media_type, category_tag, difficulty_level, sort_order)
VALUES (
    'TOEFL_S_01', 'TOEFL_S_02', 'Structure Question 2',
    'Choose the word or phrase that best completes the sentence.',
    'NONE', 'TOEFL_STRUCTURE_PART_A', 2, 2
) ON CONFLICT DO NOTHING;

INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 1, '________ by the World Health Organization, the global vaccination program successfully eradicated smallpox in 1980.', 'SINGLE_CHOICE', 1.0, 'A', 'Past participle phrase "Coordinated by..." correctly modifies the subject "the global vaccination program".'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEFL_S_02';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'Coordinated', 1),
    ('B', 'Coordinating', 2),
    ('C', 'Having coordinated', 3),
    ('D', 'It was coordinated', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEFL_S_02';

-- Question 3: Error Recognition (Parallelism / Form)
INSERT INTO cat.cat_bank_stimulus (kodesoal, stimulus_code, title, narrative_text, media_type, category_tag, difficulty_level, sort_order)
VALUES (
    'TOEFL_S_01', 'TOEFL_S_03', 'Error Recognition Question 3',
    'Identify the one underlined word or phrase that must be changed for the sentence to be correct.',
    'NONE', 'TOEFL_STRUCTURE_PART_B', 2, 3
) ON CONFLICT DO NOTHING;

INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 1, 'Modern healthcare [facilities](A) are equipped [with](B) advanced diagnostic tools to detect diseases [early](C) and [accurate](D).', 'SINGLE_CHOICE', 1.0, 'D', '"accurate" should be the adverb "accurately" to parallel the adverb "early" modifying the verb "detect".'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEFL_S_03';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'facilities', 1),
    ('B', 'with', 2),
    ('C', 'early', 3),
    ('D', 'accurate', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEFL_S_03';

-- ─────────────────────────────────────────────────────────────────────────────
-- SECTION 3: TOEFL ITP - READING COMPREHENSION (kodesoal: 'TOEFL_R_01')
-- ─────────────────────────────────────────────────────────────────────────────

-- Passage 1: The Evolution of Epidemiology and Public Health
INSERT INTO cat.cat_bank_stimulus (kodesoal, stimulus_code, title, narrative_text, media_type, category_tag, difficulty_level, sort_order)
VALUES (
    'TOEFL_R_01', 'TOEFL_R_P01', 'The Evolution of Modern Epidemiology and Disease Prevention',
    'Epidemiology, the branch of medical science that deals with the incidence, distribution, and control of disease in populations, originated during the mid-nineteenth century. Prior to this period, infectious diseases were commonly attributed to the "miasma theory"—the belief that toxic vapors emanating from decaying matter caused widespread pestilence.\n\nThe foundational breakthrough in epidemiological methodology occurred in 1854 during a severe cholera outbreak in London. Dr. John Snow, skeptical of the miasma doctrine, systematically mapped the geographical cluster of cholera fatalities in the Soho district. His meticulous cartographic investigation demonstrated that the overwhelming majority of deaths occurred in close proximity to the public water pump on Broad Street. By convincing local authorities to remove the pump handle, Snow effectively curtailed the epidemic and established waterborne contagion as the transmission vector.\n\nFollowing Snow''s pioneering work, the late nineteenth century witnessed the rapid development of the germ theory of disease, led by Louis Pasteur and Robert Koch. Modern epidemiology has since expanded far beyond infectious outbreak surveillance to encompass chronic disease etiology, environmental toxicological exposure, and genetic predispositions, utilizing complex statistical modeling to design preventative health interventions worldwide.',
    'NONE', 'TOEFL_READING', 2, 1
) ON CONFLICT DO NOTHING;

-- Question 1 on Passage 1: Main Idea
INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 1, 'What is the primary topic of the passage?', 'SINGLE_CHOICE', 1.0, 'B', 'The passage outlines the historical origin and evolution of epidemiology from early theories to modern methodology.'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEFL_R_P01';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'The microbiological mechanisms of the cholera pathogen.', 1),
    ('B', 'The historical development and scope of epidemiological science.', 2),
    ('C', 'A comparison between Dr. John Snow and Louis Pasteur.', 3),
    ('D', 'The flaws of nineteenth-century municipal water infrastructure in London.', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEFL_R_P01' AND i.item_order = 1;

-- Question 2 on Passage 1: Vocabulary in Context
INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 2, 'The word "curtailed" in paragraph 2 is closest in meaning to which of the following?', 'SINGLE_CHOICE', 1.0, 'C', '"Curtailed" means reduced, limited, or stopped.'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEFL_R_P01';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'exacerbated', 1),
    ('B', 'investigated', 2),
    ('C', 'halted', 3),
    ('D', 'prolonged', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEFL_R_P01' AND i.item_order = 2;

-- Question 3 on Passage 1: Factual Detail
INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 3, 'According to the passage, what did Dr. John Snow discover during the 1854 London cholera outbreak?', 'SINGLE_CHOICE', 1.0, 'D', 'Snow discovered that cholera fatalities were clustered around the contaminated Broad Street water pump.'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEFL_R_P01';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'Toxic vapors were the primary mode of contagion.', 1),
    ('B', 'Cholera was genetically inherited among urban residents.', 2),
    ('C', 'Antibiotics could immediately cure bacterial infections.', 3),
    ('D', 'The disease was transmitted through contaminated water from a specific pump.', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEFL_R_P01' AND i.item_order = 3;

-- ─────────────────────────────────────────────────────────────────────────────
-- SECTION 4: TOEIC - LISTENING & READING (kodesoal: 'TOEIC_L_01' & 'TOEIC_R_01')
-- ─────────────────────────────────────────────────────────────────────────────

-- TOEIC Listening: Workplace Scenario
INSERT INTO cat.cat_bank_stimulus (kodesoal, stimulus_code, title, narrative_text, media_type, media_url, category_tag, difficulty_level, sort_order)
VALUES (
    'TOEIC_L_01', 'TOEIC_L_01', 'Workplace Briefing: Medical Equipment Procurement',
    'Attention all hospital department heads. The annual procurement committee meeting scheduled for Thursday has been moved to Friday at 2:00 PM in Conference Room B. Please ensure that your quarterly inventory audits and equipment replacement requisitions are submitted to the administrative office by Wednesday afternoon.',
    'AUDIO', 'https://actions.google.com/sounds/v1/human_voices/human_voice_male_dialogue.ogg',
    'TOEIC_LISTENING', 2, 1
) ON CONFLICT DO NOTHING;

INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 1, 'When will the procurement committee meeting take place?', 'SINGLE_CHOICE', 1.0, 'B', 'The speaker mentions that the meeting has been moved to Friday at 2:00 PM.'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEIC_L_01';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'Wednesday afternoon', 1),
    ('B', 'Friday at 2:00 PM', 2),
    ('C', 'Thursday morning', 3),
    ('D', 'Next Monday at 9:00 AM', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEIC_L_01';

-- TOEIC Reading: Business Email / Schedule Notice
INSERT INTO cat.cat_bank_stimulus (kodesoal, stimulus_code, title, narrative_text, media_type, category_tag, difficulty_level, sort_order)
VALUES (
    'TOEIC_R_01', 'TOEIC_R_01', 'Memo: Healthcare Accreditation Workshop',
    'MEMORANDUM\nTO: All Clinical Staff & Department Supervisors\nFROM: Director of Quality Assurance\nDATE: August 28, 2026\nSUBJECT: ISO 9001 Hospital Accreditation Workshop\n\nIn preparation for our upcoming national healthcare accreditation inspection, a mandatory two-day training seminar will be conducted on September 15-16 in the Main Auditorium. All healthcare practitioners are requested to review the standardized clinical documentation guidelines available on the hospital intranet prior to attending.',
    'NONE', 'TOEIC_READING', 2, 1
) ON CONFLICT DO NOTHING;

INSERT INTO cat.cat_stimulus_items (id_stimulus, item_order, question_text, item_type, weight_correct, correct_answer, explanation)
SELECT id_stimulus, 1, 'What are staff members instructed to do before attending the workshop?', 'SINGLE_CHOICE', 1.0, 'A', 'Staff are asked to review the standardized clinical documentation guidelines on the intranet.'
FROM cat.cat_bank_stimulus WHERE stimulus_code = 'TOEIC_R_01';

INSERT INTO cat.cat_item_options (id_item, option_label, option_text, sort_order)
SELECT id_item, o.label, o.text, o.sorder
FROM cat.cat_stimulus_items i
JOIN cat.cat_bank_stimulus s ON s.id_stimulus = i.id_stimulus
CROSS JOIN (VALUES
    ('A', 'Review clinical documentation guidelines on the intranet.', 1),
    ('B', 'Submit a formal application to the Quality Assurance director.', 2),
    ('C', 'Register for an external training certification program.', 3),
    ('D', 'Conduct an independent audit of patient medical records.', 4)
) AS o(label, text, sorder)
WHERE s.stimulus_code = 'TOEIC_R_01';
