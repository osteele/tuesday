# Generate a matrix of Ruby strftime results for differential Go tests.

require 'csv'
require 'date'
require 'time'

ENV['LC_ALL'] = 'C'
ENV['TZ'] = 'UTC'

sources = [
  '1600-01-01T00:00:00.000000001+00:00',
  '1969-12-31T23:59:59.999500000+00:00',
  '1970-01-01T00:00:00.000000000+00:00',
  '2006-01-02T15:04:05.123456789-05:00',
  '2017-01-01T00:00:00.000000001+05:30',
  '3000-12-31T23:59:59.999999999+00:00',
]

conversions = %w[A B C D F G H I L M N P R S T U V W X Y a b c d e g h j k l m n p r s t u v w x y z %]
formats = conversions.map { |conversion| "%#{conversion}" }
formats.concat %w[
  %10A %010A %_10A %-10A %-^A %#A %#^A
  %1m %3m %-3m %_3m %03m %_0m %0_m
  %1N %3N %6N %9N %12N %-N %_N
  %1L %3L %6L %9L %12L %-L %_L
  %_z %10z %-10z %:z %::z %:::z %10:z %_10:z
  %#c %^c %#^c
  %J %_J %EJ %:J
]
formats.uniq!

CSV($stdout) do |csv|
  sources.each do |source|
    value = Time.iso8601(source)
    formats.each do |format|
      csv << [source, format, value.strftime(format)]
    end

    datetime = DateTime.iso8601(source)
    %w[%Q %1Q %_Q].each do |format|
      csv << [source, format, datetime.strftime(format)]
    end
  end
end
