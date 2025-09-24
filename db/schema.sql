CREATE TABLE composers (
    composer_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    short_name VARCHAR(100) NOT NULL,
    full_name VARCHAR(100) NOT NULL
);

CREATE TABLE pieces (
    piece_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    composer_id INT REFERENCES composers(composer_id) ON DELETE CASCADE NOT NULL,
    piece_title VARCHAR(100) NOT NULL
);

CREATE TABLE programmes (
    programme_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    programme_title VARCHAR(100) NOT NULL
);

CREATE TABLE programme_pieces (
    programme_id INT REFERENCES programmes(programme_id) ON DELETE CASCADE,
    piece_id INT REFERENCES pieces(piece_id) ON DELETE CASCADE,
    sequence INT NOT NULL CHECK (sequence > 0),
    PRIMARY KEY (programme_id, piece_id)
);

CREATE TABLE venues (
    venue_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    address VARCHAR(100) NOT NULL
);

CREATE TABLE events (
    event_id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_title VARCHAR(100) NOT NULL,
    venue_id INT REFERENCES venues(venue_id),
    event_date TIMESTAMP NOT NULL,
    ticket_link VARCHAR(255),
    programme_id INT REFERENCES programmes(programme_id) ON DELETE CASCADE NOT NULL
);
