class PostsController < ApplicationController
  # Insecure: Skips authorization checks for all actions, exposing the app to unauthorized access.
  # <expect-error> Skipping authorization for all actions
  skip_authorization_check

  def index
    @posts = Post.all
    render json: @posts
  end

  def show
    @post = Post.find(params[:id])
    render json: @post
  end
end


class PostsController < ApplicationController
  # Safe Ensure authorization checks are performed for all actions.
  before_action :authenticate_user!      # Ensures only logged-in users can access.
  load_and_authorize_resource           # CanCanCan method to load resource and check permissions.

  def index
    # @posts is automatically loaded and authorized by load_and_authorize_resource.
    render json: @posts
  end

  def show
    render json: @post
  end
end
