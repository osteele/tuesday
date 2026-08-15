#!/usr/bin/env ruby
# Reports the strftime behavior of the Ruby runtime that executes it.
# Used to compare Ruby 2.7 (GitHub Pages / Jekyll) against Ruby 3.4.

require 'time'
require 'date'
require 'json'

ENV['LC_ALL'] = 'C'

SOURCES = [
  '2006-01-02T15:04:05.123456789-05:00',
  '2006-01-02T15:04:05.123456789+05:30',
  '2006-07-15T15:04:05.123456789+00:00',
  '1969-12-31T23:59:59.999500000+00:00',
].freeze

FLAGS = ['', '-', '_', '^', '#', '0', '-_', '_0', '0_', '-0', '0-', '^#', '#^'].freeze

# Core conversions plus the composite and special forms.
CONVERSIONS = (
  ('A'..'Z').to_a + ('a'..'z').to_a + %w[+ % n t]
).freeze

SPECIAL_ZONE = %w[z :z ::z :::z].freeze

WIDTHS = [1, 2, 3, 5, 10].freeze

def build_formats
  formats = []

  CONVERSIONS.each do |c|
    FLAGS.each do |f|
      formats << "%#{f}#{c}"
    end
  end

  SPECIAL_ZONE.each do |c|
    FLAGS.each do |f|
      formats << "%#{f}#{c}"
    end
  end

  # Width variations for a representative set of conversions.
  %w[m e H k I l d j Y y C z :z ::z :::z A Z B b p P].each do |c|
    WIDTHS.each do |w|
      formats << "%#{w}#{c}"
      formats << "%0#{w}#{c}"
      formats << "%_#{w}#{c}"
      formats << "%-#{w}#{c}"
    end
  end

  formats.uniq
end

def evaluate(time_value, formats)
  result = {}
  formats.each do |fmt|
    begin
      result[fmt] = time_value.strftime(fmt)
    rescue StandardError => e
      result[fmt] = { error: e.class.to_s, message: e.message }
    end
  end
  result
end

report = {
  ruby_version: RUBY_VERSION,
  ruby_description: RUBY_DESCRIPTION,
  generated_at: Time.now.utc.iso8601,
  note: 'This report captures the strftime behavior of the Ruby runtime it is executed on.',
  times: []
}

formats = build_formats

SOURCES.each do |source|
  time_value = Time.iso8601(source)
  report[:times] << {
    source: source,
    iso8601: time_value.iso8601(9),
    zone_name: time_value.zone,
    zone_offset_seconds: time_value.utc_offset,
    formats: evaluate(time_value, formats)
  }
end

# DateTime-only %Q coverage.
datetime = DateTime.iso8601(SOURCES.first)
report[:datetime] = {
  source: SOURCES.first,
  iso8601: datetime.iso8601(9),
  formats: evaluate(datetime, FLAGS.map { |f| "%#{f}Q" })
}

puts JSON.pretty_generate(report)
