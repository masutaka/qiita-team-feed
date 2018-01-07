require 'qiita'
require 'rss'

client = Qiita::Client.new(
  access_token: ENV['QIITA_ACCESS_TOKEN'],
  team: ENV['QIITA_TEAM_NAME']
)

items = client.list_items(per_page: 50)

if items.status != 200
  exit 1
end

def content(user_id:, user_icon_url:)
<<EOC
<a href="/#{user_id}/items" rel="noreferrer">
  <img alt="@#{user_id}" width="32" height="32" src="#{user_icon_url}">
</a>
EOC
end

atom = RSS::Maker.make('atom') do |maker|
  maker.channel.about = "https://#{ENV['QIITA_TEAM_NAME']}.qiita.com"
  maker.channel.title = "#{ENV['QIITA_TEAM_NAME']} Qiita:Team"
  maker.channel.link = "https://#{ENV['QIITA_TEAM_NAME']}.qiita.com"

  maker.channel.author = ENV['QIITA_TEAM_NAME']
  maker.channel.date = Time.now

  maker.items.do_sort = true

  items.body.each do |qitem|
    maker.items.new_item do |aitem|
      aitem.link = qitem['url']
      aitem.title = '%s by @%s' % [qitem['title'], qitem['user']['id']]
      aitem.date = Time.parse(qitem['created_at'])
      aitem.author = qitem['user']['id']
      aitem.content.type = 'html'
      aitem.content.content = content(
        user_id: qitem['user']['id'],
        user_icon_url: qitem['user']['profile_image_url']
      )
    end
  end
end

File.write("#{ENV['SECRET']}.atom", atom)
