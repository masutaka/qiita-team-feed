require 'qiita'
require "rss"

client = Qiita::Client.new(
  access_token: ENV['QIITA_ACCESS_TOKEN'],
  team: 'feedforce'
)

items = client.list_items

if items.status != 200
  exit 1
end

# before_etag = File.read('etag.txt') rescue nil
# 
# if before_etag == items.headers['Etag']
#   exit 0
# end
# 
# File.write('etag.txt', items.headers['Etag'])

atom = RSS::Maker.make('atom') do |maker|
  maker.channel.about = "https://masutaka.net/#{ENV['SECRET']}.atom"
  maker.channel.title = 'Feedforce Qiita:Team'
  maker.channel.link = 'https://feedforce.qiita.com/'

  maker.channel.author = 'Feedforce Inc.'
  maker.channel.date = Time.now

  maker.items.do_sort = true

  items.body.each do |qitem|
    maker.items.new_item do |aitem|
      aitem.link = qitem['url']
      aitem.title = '%s by @%s' % [qitem['title'], qitem['user']['id']]
      aitem.date = Time.parse(qitem['created_at'])
    end
  end
end

File.write("#{ENV['SECRET']}.atom", atom)
